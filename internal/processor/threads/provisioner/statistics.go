package provisioner

import (
	"github.com/Sentmint/PipelineOps/internal/processor/provision/pipeline"
)

func (thread *Thread) getStatistics() []*pipeline.Pipeline {

	statistics := make([]*pipeline.Pipeline, 0)

	for _, s := range thread.provisioner.GetSupervisors() {

		statistics = append(statistics, s.Pipeline)
	}

	return statistics
}
