package cluster

import (
	"ETLFramework/channel"
	"ETLFramework/net"
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
	supervisor.Config = NewConfig(net.EmptyString, DefaultChannelThreshold, DefaultChannelGrowthFactor, DefaultChannelThreshold, DefaultChannelGrowthFactor, DoNothing)
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

func (m *Supervisor) Start() *Response {
	m.waitGroup.Add(3)

	startTime := time.Now()

	// start creating the default frontend goroutines
	m.Provision(Extract)
	m.Provision(Transform)
	m.Provision(Load)
	// end creating the default frontend goroutines

	// every N seconds we should check if the etChannel or tlChannel is congested
	// and requires us to provision additional nodes
	go m.Runtime()

	m.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	response := NewResponse(m.Config, m.Stats, time.Now().Sub(startTime))
	return response
}

func (m *Supervisor) Runtime() {
	for {
		// is etChannel congested?
		if m.etChannel.State == channel.Congested {
			n := m.Stats.NumProvisionedTransformRoutes
			for n > 0 {
				m.Provision(Transform)
				n--
			}
			m.Stats.NumProvisionedTransformRoutes *= m.etChannel.Config.GrowthFactor
		}
		// is tlChannel congested?
		if m.tlChannel.State == channel.Congested {
			n := m.Stats.NumProvisionedLoadRoutines
			for n > 0 {
				m.Provision(Load)
				n--
			}
			m.Stats.NumProvisionedLoadRoutines *= m.tlChannel.Config.GrowthFactor
		}

		// check if the channel is congested after DefaultMonitorRefreshDuration seconds
		time.Sleep(DefaultMonitorRefreshDuration * time.Second)
	}
}

func (m *Supervisor) Provision(segment Segment) {
	go func() {
		switch segment {
		case Extract:
			m.group.ExtractFunc(m.etChannel.Channel)
			break
		case Transform: // transform
			m.Stats.NumProvisionedTransformRoutes++
			m.group.TransformFunc(m.etChannel.Channel, m.tlChannel.Channel)
			break
		default: // load
			m.Stats.NumProvisionedTransformRoutes++
			m.group.LoadFunc(m.tlChannel.Channel)
			break
		}
		m.waitGroup.Done() // notify the wait group a process has completed ~ if all are finished we close the monitor
	}()
}
