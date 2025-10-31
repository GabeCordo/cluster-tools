package in_memory

import (
	"github.com/FortifiedCode/flock/internal/core/database/pipeline"
	"testing"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalPipelineDatabase()

	var i pipeline.Database
	i = d

	i.Print()
}
