package cluster

import (
	"ETLFramework/channel"
	"time"
)

const (
	DefaultMonitorRefreshDuration = 1
)

func NewMonitor(cluster Cluster) *Monitor {
	monitor := new(Monitor)

	monitor.group = cluster
	monitor.Config = Config{"", 10, 2, 10, 2}
	monitor.etChannel = channel.NewManagedChannel(monitor.Config.etChannelThreshold, monitor.Config.etChannelGrowthFactor)
	monitor.tlChannel = channel.NewManagedChannel(monitor.Config.tlChannelThreshold, monitor.Config.tlChannelGrowthFactor)

	return monitor
}

func NewCustomMonitor(cluster Cluster, config Config) *Monitor {
	monitor := new(Monitor)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	monitor.group = cluster
	monitor.Config = config
	monitor.etChannel = channel.NewManagedChannel(config.etChannelThreshold, config.etChannelGrowthFactor)
	monitor.tlChannel = channel.NewManagedChannel(config.tlChannelThreshold, config.tlChannelGrowthFactor)

	return monitor
}

func (m *Monitor) Start() Response {
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

	response := Response{Stats: m.Stats, LapsedTime: time.Now().Sub(startTime)}
	return response
}

func (m *Monitor) Runtime() {
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

func (m *Monitor) Provision(segment Segment) {
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
