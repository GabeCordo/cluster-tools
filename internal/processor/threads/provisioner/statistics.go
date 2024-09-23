package provisioner

import (
	"github.com/GabeCordo/cluster-tools/internal/processor/supervisor"
)

func (thread *Thread) getStatistics() []*supervisor.Summary {

	defer thread.requestWg.Done()

	statistics := make([]*supervisor.Summary, 0)

	for _, s := range thread.provisioner.GetSupervisors() {

		summary := new(supervisor.Summary)

		// TODO : fix
		//summary.Namespace = s.
		//summary.Cluster = cluster.Identifier
		summary.Supervisor = s.Id
		summary.Statistics = s.Stats

		statistics = append(statistics, summary)
	}

	return statistics
}
