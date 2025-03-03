package provisioner

import (
	"github.com/Sentmint/yule"
)

func (thread *Thread) getStatistics() []*yule.Statistics {

	statistics := make([]*yule.Statistics, 0)

	for _, r := range thread.runnablePipelines {

		statistics = append(statistics, r.Stats)
	}

	return statistics
}
