package supervisor

import (
	"fmt"
	"github.com/GabeCordo/etl/components/channel"
	"github.com/GabeCordo/etl/components/cluster"
	"log"
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
	supervisor.ETChannel = channel.NewManagedChannel("ETChannel", supervisor.Config.ETChannelThreshold, supervisor.Config.ETChannelGrowthFactor)
	supervisor.TLChannel = channel.NewManagedChannel("TLChannel", supervisor.Config.TLChannelThreshold, supervisor.Config.TLChannelGrowthFactor)

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
	supervisor.ETChannel = channel.NewManagedChannel("ETChannel", config.ETChannelThreshold, config.ETChannelGrowthFactor)
	supervisor.TLChannel = channel.NewManagedChannel("TLChannel", config.TLChannelThreshold, config.TLChannelGrowthFactor)

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

	// every N seconds we should check if the ETChannel or TLChannel is congested
	// and requires us to provision additional nodes
	go supervisor.Runtime()

	supervisor.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	// calculate the timings produced by data being fed across each of the channels
	supervisor.CalculateTiming()

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
		supervisor.ETChannel.GetState()

		// is ETChannel congested?
		if supervisor.ETChannel.GetState() == channel.Congested {
			supervisor.Stats.Channels.NumEtThresholdBreaches++
			n := supervisor.Stats.Threads.NumProvisionedTransformRoutes
			for n > 0 {
				supervisor.Provision(cluster.Transform)
				n--
			}
			supervisor.Stats.Threads.NumProvisionedTransformRoutes *= supervisor.ETChannel.GetGrowthFactor()
		}

		supervisor.TLChannel.GetState()

		// is TLChannel congested?
		if supervisor.TLChannel.GetState() == channel.Congested {
			supervisor.Stats.Channels.NumTlThresholdBreaches++
			n := supervisor.Stats.Threads.NumProvisionedLoadRoutines
			for n > 0 {
				supervisor.Provision(cluster.Load)
				n--
			}
			supervisor.Stats.Threads.NumProvisionedLoadRoutines *= supervisor.TLChannel.GetGrowthFactor()
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
				defer func() {
					if r := recover(); r != nil {
						log.Println("cluster.Extract function raised error")
						supervisor.ETChannel.ProducerDone()
						supervisor.waitGroup.Done()
					}
				}()
				supervisor.Stats.Threads.NumProvisionedExtractRoutines++
				oneWayChannel, _ := channel.NewOneWayManagedChannel(supervisor.ETChannel)

				supervisor.ETChannel.AddProducer()
				supervisor.group.ExtractFunc(oneWayChannel)
				supervisor.ETChannel.ProducerDone()
			}
		case cluster.Transform:
			{
				defer func() {
					if r := recover(); r != nil {
						log.Println("cluster.Transform function raised error")
						log.Println(r)
						supervisor.TLChannel.ProducerDone()
						supervisor.waitGroup.Done()
					}
				}()

				supervisor.Stats.Threads.NumProvisionedTransformRoutes++
				supervisor.TLChannel.AddProducer()

				fmt.Println("starting to read ET Channel")

				for request := range supervisor.ETChannel.GetChannel() {

					// associates a TimeOut to the data being removed from the channel and decrements
					// the data counter for the current pipe
					supervisor.ETChannel.DataPopped(request.Id)

					supervisor.Stats.Data.TotalOverETChannel++

					if i, ok := (supervisor.group).(cluster.VerifiableET); ok && !i.VerifyETFunction(request) {
						continue
					}

					data, success := supervisor.group.TransformFunc(request.Data)
					if success {
						supervisor.Stats.Data.TotalOverTLChannel++
						supervisor.TLChannel.Push(channel.DataWrapper{Id: request.Id, Data: data})
					}
				}

				supervisor.TLChannel.ProducerDone()
			}
		case cluster.Load:
			{
				defer func() {
					if r := recover(); r != nil {
						log.Println("cluster.Load function raised error")
						supervisor.waitGroup.Done()
					}
				}()

				supervisor.Stats.Threads.NumProvisionedLoadRoutines++

				for request := range supervisor.TLChannel.GetChannel() {
					supervisor.Stats.Data.TotalProcessed++

					// associates a TimeOut to the data being removed from the channel and decrements
					// the data counter for the current pipe
					supervisor.TLChannel.DataPopped(request.Id)

					if i, ok := (supervisor.group).(cluster.VerifiableTL); ok && !i.VerifyTLFunction(request) {
						continue
					}

					supervisor.group.LoadFunc(request.Data)
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

func (supervisor *Supervisor) Deletable() bool {
	return (supervisor.State == Terminated) || (supervisor.State == Failed)
}

func (supervisor *Supervisor) CalculateTiming() {

	supervisor.Stats.Data.TotalDropped = supervisor.Stats.Data.TotalOverETChannel - supervisor.Stats.Data.TotalOverTLChannel
	supervisor.Stats.CalculateTiming(supervisor.ETChannel.Timestamps, supervisor.TLChannel.Timestamps)
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
