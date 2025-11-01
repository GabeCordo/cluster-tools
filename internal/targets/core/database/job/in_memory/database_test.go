package in_memory

import (
	"github.com/FortifiedCode/flock/internal/targets/core/database/job"
	"testing"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalJobDatabase()

	var i job.Database
	i = d

	i.Print()
}
