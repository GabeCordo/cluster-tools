package in_memory

import (
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/pipeline"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalPipelineDatabase()

	var i pipeline.Database
	i = d

	i.Print()
}
