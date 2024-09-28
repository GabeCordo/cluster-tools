package provisioner

import (
	"github.com/GabeCordo/cluster-tools/internal/processor/provision/pipeline"
)

func (thread *Thread) getStatistics() []*pipeline.Pipeline {

	defer thread.requestWg.Done()

	statistics := make([]*pipeline.Pipeline, 0)

	for _, s := range thread.provisioner.GetSupervisors() {

		statistics = append(statistics, s.Pipeline)
	}

	return statistics
}
