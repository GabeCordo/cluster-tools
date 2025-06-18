package provisioner

import (
	"github.com/GabeCordo/Flock/internal/processor/component/provision/pipeline"
)

func (t *Thread) getStatistics() []*pipeline.Pipeline {

	statistics := make([]*pipeline.Pipeline, 0)

	for _, s := range t.provisioner.GetSupervisors() {

		statistics = append(statistics, s.Pipeline)
	}

	return statistics
}
