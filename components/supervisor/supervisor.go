package supervisor

import (
	"fmt"
	"github.com/GabeCordo/etl/components/channel"
	"github.com/GabeCordo/etl/components/cluster"
	"time"
)

const (
	DefaultNumberOfClusters       = 1
	DefaultMonitorRefreshDuration = 1
	DefaultChannelThreshold       = 10
	DefaultChannelGrowthFactor    = 2
)

func NewSupervisor(clusterName string, clusterImplementation cluster.Cluster) *Supervisor {
	supervisor := new(Supervisor)

	supervisor.group = clusterImplementation
	supervisor.Config = cluster.Config{
		clusterName,
		DefaultChannelThreshold,
		DefaultNumberOfClusters,
		DefaultNumberOfClusters,
		DefaultChannelGrowthFactor,
		DefaultChannelThreshold,
		DefaultChannelGrowthFactor,
		DefaultChannelGrowthFactor,
	}
	supervisor.Stats = cluster.NewStatistics()
	supervisor.etChannel = channel.NewManagedChannel("etChannel", supervisor.Config.ETChannelThreshold, supervisor.Config.ETChannelGrowthFactor)
	supervisor.tlChannel = channel.NewManagedChannel("tlChannel", supervisor.Config.TLChannelThreshold, supervisor.Config.TLChannelGrowthFactor)

	return supervisor
}

func NewCustomSupervisor(clusterImplementation cluster.Cluster, config cluster.Config) *Supervisor {
	supervisor := new(Supervisor)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	supervisor.group = clusterImplementation
	supervisor.Config = config
	supervisor.Stats = cluster.NewStatistics()
	supervisor.etChannel = channel.NewManagedChannel("etChannel", config.ETChannelThreshold, config.ETChannelGrowthFactor)
	supervisor.tlChannel = channel.NewManagedChannel("tlChannel", config.TLChannelThreshold, config.TLChannelGrowthFactor)

	return supervisor
}

func (supervisor *Supervisor) Event(event Event) bool {
	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()

	if supervisor.State == UnTouched {
		if event == Startup {
			supervisor.State = Running
		} else {
			return false
		}
	} else if supervisor.State == Running {
		if event == StartProvision {
			supervisor.State = Provisioning
		} else if event == Error {
			supervisor.State = Failed
		} else if event == TearedDown {
			supervisor.State = Terminated
		} else {
			return false
		}
	} else if supervisor.State == Provisioning {
		if event == EndProvision {
			supervisor.State = Running
		} else if event == Error {
			supervisor.State = Failed
		} else {
			return false
		}
	} else if (supervisor.State == Failed) || (supervisor.State == Terminated) {
		return false
	}

	return true // represents a boolean ~ hasStateChanged?
}

func (supervisor *Supervisor) Start() (response *cluster.Response) {
	supervisor.Event(Startup)

	defer supervisor.Event(TearedDown)

	defer func() {
		// has the user defined function crashed during runtime?
		if r := recover(); r != nil {
			// yes => return a response that identifies that the cluster crashed
			response = cluster.NewResponse(
				supervisor.Config,
				supervisor.Stats,
				time.Now().Sub(supervisor.StartTime),
				true,
			)
		}
	}()

	supervisor.StartTime = time.Now()

	//// start creating the default frontend goroutines
	supervisor.Provision(cluster.Extract)

	// the config specifies the number of transform functions to start running in parallel
	for i := 0; i < supervisor.Config.StartWithNTransformClusters; i++ {
		supervisor.Provision(cluster.Transform)
	}

	// the config specifies the number of load functions to start running in parallel
	for i := 0; i < supervisor.Config.StartWithNLoadClusters; i++ {
		supervisor.Provision(cluster.Load)
	}

	//// end creating the default frontend goroutines

	// every N seconds we should check if the etChannel or tlChannel is congested
	// and requires us to provision additional nodes
	go supervisor.Runtime()

	supervisor.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	response = cluster.NewResponse(
		supervisor.Config,
		supervisor.Stats,
		time.Now().Sub(supervisor.StartTime),
		false,
	)

	return response
}

func (supervisor *Supervisor) Runtime() {
	for {
		supervisor.etChannel.GetState()

		// is etChannel congested?
		if supervisor.etChannel.GetState() == channel.Congested {
			supervisor.Stats.NumEtThresholdBreaches++
			n := supervisor.Stats.NumProvisionedTransformRoutes
			for n > 0 {
				supervisor.Provision(cluster.Transform)
				n--
			}
			supervisor.Stats.NumProvisionedTransformRoutes *= supervisor.etChannel.GetGrowthFactor()
		}

		supervisor.tlChannel.GetState()

		// is tlChannel congested?
		if supervisor.tlChannel.GetState() == channel.Congested {
			supervisor.Stats.NumTlThresholdBreaches++
			n := supervisor.Stats.NumProvisionedLoadRoutines
			for n > 0 {
				supervisor.Provision(cluster.Load)
				n--
			}
			supervisor.Stats.NumProvisionedLoadRoutines *= supervisor.tlChannel.GetGrowthFactor()
		}

		// check if the channel is congested after DefaultMonitorRefreshDuration seconds
		time.Sleep(DefaultMonitorRefreshDuration * time.Second)
	}
}

func (supervisor *Supervisor) Provision(segment cluster.Segment) {
	supervisor.Event(StartProvision)
	defer supervisor.Event(EndProvision)

	go func() {
		switch segment {
		case cluster.Extract:
			{
				supervisor.Stats.NumProvisionedExtractRoutines++
				oneWayChannel, _ := channel.NewOneWayManagedChannel(supervisor.etChannel)

				supervisor.etChannel.AddProducer()
				supervisor.group.ExtractFunc(oneWayChannel)
				supervisor.etChannel.ProducerDone()
			}
		case cluster.Transform:
			{
				supervisor.Stats.NumProvisionedTransformRoutes++
				supervisor.tlChannel.AddProducer()

				for request := range supervisor.etChannel.Channel {
					supervisor.etChannel.Pull()

					if i, ok := (supervisor.group).(cluster.VerifiableET); ok && !i.VerifyETFunction(request) {
						continue
					}

					data := supervisor.group.TransformFunc(request)
					supervisor.tlChannel.Push(data)
				}

				supervisor.tlChannel.ProducerDone()
			}
		case cluster.Load:
			{
				supervisor.Stats.NumProvisionedLoadRoutines++

				for request := range supervisor.tlChannel.Channel {
					supervisor.Stats.NumOfDataProcessed++
					supervisor.tlChannel.Pull()

					if i, ok := (supervisor.group).(cluster.VerifiableTL); ok && !i.VerifyTLFunction(request) {
						continue
					}

					supervisor.group.LoadFunc(request)
				}
			}
		}

		// notify the wait group a process has completed ~ if all are finished we close the monitor
		supervisor.waitGroup.Done()
	}()

	// a new function (E, T, or L) is provisioned
	// we should inform the wait group that the supervisor isn't finished until the wg is done
	supervisor.waitGroup.Add(1)
}

func (supervisor *Supervisor) Print() {
	fmt.Printf("Id: %d\n", supervisor.Id)
	fmt.Printf("Cluster: %s\n", supervisor.Config.Identifier)
}

func (status Status) ToString() string {
	switch status {
	case UnTouched:
		return "UnTouched"
	case Running:
		return "Running"
	case Provisioning:
		return "Provisioning"
	case Failed:
		return "Failed"
	case Terminated:
		return "Terminated"
	default:
		return "None"
	}
}
