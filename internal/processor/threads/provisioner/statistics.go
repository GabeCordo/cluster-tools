package provisioner

import "github.com/GabeCordo/cluster-tools/internal/processor/interfaces"

func (thread *Thread) getStatistics() []*interfaces.SupervisorSummary {

	defer thread.requestWg.Done()

	modules := GetProvisionerInstance().GetModules()

	statistics := make([]*interfaces.SupervisorSummary, 0)

	for _, module := range modules {

		for _, cluster := range module.GetClusters() {

			for _, supervisor := range cluster.FindSupervisors() {

				summary := new(interfaces.SupervisorSummary)

				summary.Module = module.Identifier
				summary.Cluster = cluster.Identifier
				summary.Supervisor = supervisor.Id
				summary.Statistics = supervisor.Stats
				summary.ETState = supervisor.ETChannel.State.ToString()
				summary.ETSize = supervisor.ETChannel.Size
				summary.TLState = supervisor.TLChannel.State.ToString()
				summary.TLSize = supervisor.TLChannel.Size

				statistics = append(statistics, summary)
			}
		}
	}

	return statistics
}
