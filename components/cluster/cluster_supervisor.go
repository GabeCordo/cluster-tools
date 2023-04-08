package cluster

import (
	"github.com/GabeCordo/etl/components/channel"
	"github.com/GabeCordo/fack"
	"time"
)

const (
	DefaultMonitorRefreshDuration = 1
	DefaultChannelThreshold       = 10
	DefaultChannelGrowthFactor    = 2
)

func NewSupervisor(cluster Cluster) *Supervisor {
	supervisor := new(Supervisor)

	supervisor.group = cluster
	supervisor.Config = NewConfig(fack.EmptyString, DefaultChannelThreshold, DefaultChannelGrowthFactor, DefaultChannelThreshold, DefaultChannelGrowthFactor, DoNothing)
	supervisor.Stats = NewStatistics()
	supervisor.etChannel = channel.NewManagedChannel(supervisor.Config.ETChannelThreshold, supervisor.Config.ETChannelGrowthFactor)
	supervisor.tlChannel = channel.NewManagedChannel(supervisor.Config.TLChannelThreshold, supervisor.Config.TLChannelGrowthFactor)

	return supervisor
}

func NewCustomSupervisor(cluster Cluster, config *Config) *Supervisor {
	supervisor := new(Supervisor)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	supervisor.group = cluster
	supervisor.Config = config
	supervisor.Stats = NewStatistics()
	supervisor.etChannel = channel.NewManagedChannel(config.ETChannelThreshold, config.ETChannelGrowthFactor)
	supervisor.tlChannel = channel.NewManagedChannel(config.TLChannelThreshold, config.TLChannelGrowthFactor)

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

func (supervisor *Supervisor) Start() *Response {
	supervisor.Event(Startup)
	defer supervisor.Event(TearedDown)

	supervisor.waitGroup.Add(3)

	supervisor.StartTime = time.Now()

	// start creating the default frontend goroutines
	supervisor.Provision(Extract)
	supervisor.Provision(Transform)
	supervisor.Provision(Load)
	// end creating the default frontend goroutines

	// every N seconds we should check if the etChannel or tlChannel is congested
	// and requires us to provision additional nodes
	go supervisor.Runtime()

	supervisor.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	response := NewResponse(supervisor.Config, supervisor.Stats, time.Now().Sub(supervisor.StartTime))
	return response
}

func (supervisor *Supervisor) Runtime() {
	for {
		// is etChannel congested?
		if supervisor.etChannel.State == channel.Congested {
			supervisor.Stats.NumEtThresholdBreaches++
			n := supervisor.Stats.NumProvisionedTransformRoutes
			for n > 0 {
				supervisor.Provision(Transform)
				n--
			}
			supervisor.Stats.NumProvisionedTransformRoutes *= supervisor.etChannel.Config.GrowthFactor
		}

		// is tlChannel congested?
		if supervisor.tlChannel.State == channel.Congested {
			supervisor.Stats.NumTlThresholdBreaches++
			n := supervisor.Stats.NumProvisionedLoadRoutines
			for n > 0 {
				supervisor.Provision(Load)
				n--
			}
			supervisor.Stats.NumProvisionedLoadRoutines *= supervisor.tlChannel.Config.GrowthFactor
		}

		// check if the channel is congested after DefaultMonitorRefreshDuration seconds
		time.Sleep(DefaultMonitorRefreshDuration * time.Second)
	}
}

func (supervisor *Supervisor) Provision(segment Segment) {
	supervisor.Event(StartProvision)
	defer supervisor.Event(EndProvision)

	go func() {
		switch segment {
		case Extract:
			supervisor.Stats.NumProvisionedExtractRoutines++
			supervisor.group.ExtractFunc(supervisor.etChannel.Channel)
			break
		case Transform: // transform
			supervisor.Stats.NumProvisionedTransformRoutes++
			supervisor.group.TransformFunc(supervisor.etChannel.Channel, supervisor.tlChannel.Channel)
			break
		default: // load
			supervisor.Stats.NumProvisionedLoadRoutines++
			supervisor.group.LoadFunc(supervisor.tlChannel.Channel)
			break
		}
		supervisor.waitGroup.Done() // notify the wait group a process has completed ~ if all are finished we close the monitor
	}()
}

func (status Status) String() string {
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
