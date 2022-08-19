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
	DefaultOnCrash                = DoNothing
)

func NewMonitor(cluster Cluster, mode ...OnCrash) *Monitor {
	monitor := new(Monitor)

	monitor.group = cluster
	monitor.Config = NewConfig(net.EmptyString, DefaultChannelThreshold, DefaultChannelGrowthFactor, DefaultChannelThreshold, DefaultChannelGrowthFactor, DefaultOnCrash)
	monitor.Stats = NewStatistics()
	monitor.etChannel = channel.NewManagedChannel(monitor.Config.ETChannelThreshold, monitor.Config.ETChannelGrowthFactor)
	monitor.tlChannel = channel.NewManagedChannel(monitor.Config.TLChannelThreshold, monitor.Config.TLChannelGrowthFactor)

	if len(mode) > 0 {
		monitor.mode = mode[0]
	} else {
		monitor.mode = DoNothing
	}
	return monitor
}

func NewCustomMonitor(cluster Cluster, config *Config) *Monitor {
	monitor := new(Monitor)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	monitor.group = cluster
	monitor.Config = config
	monitor.Stats = NewStatistics()
	monitor.etChannel = channel.NewManagedChannel(config.ETChannelThreshold, config.ETChannelGrowthFactor)
	monitor.tlChannel = channel.NewManagedChannel(config.TLChannelThreshold, config.TLChannelGrowthFactor)
	monitor.mode = config.Mode

	return monitor
}

func (m *Monitor) Start() *Response {
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

	// wait for the Extract-Transform-Load (ETL) Cycle to Complete
	m.waitGroup.Wait()

	response := NewResponse(m.Config, m.Stats, time.Now().Sub(startTime))
	return response
}

func (m *Monitor) Runtime() {
	for {
		// is etChannel congested?
		if m.etChannel.State == channel.Congested {
			m.Stats.NumEtThresholdBreaches++

			n := m.Stats.NumProvisionedTransformRoutes
			for n > 0 {
				m.Provision(Transform)
				n--
			}
			m.Stats.NumProvisionedTransformRoutes *= m.etChannel.Config.GrowthFactor
		}
		// is tlChannel congested?
		if m.tlChannel.State == channel.Congested {
			m.Stats.NumTlThresholdBreaches++

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
			m.Stats.NumProvisionedExtractRoutines++
			m.group.ExtractFunc(m.etChannel.Channel)
			break
		case Transform:
			m.Stats.NumProvisionedTransformRoutes++
			m.group.TransformFunc(m.etChannel.Channel, m.tlChannel.Channel)
			break
		default:
			m.Stats.NumProvisionedLoadRoutines++
			m.group.LoadFunc(m.tlChannel.Channel)
			break
		}
		m.waitGroup.Done() // notify the wait group a process has completed ~ if all are finished we close the monitor
	}()
}
