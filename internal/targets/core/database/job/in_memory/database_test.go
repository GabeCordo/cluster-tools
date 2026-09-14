package in_memory

import (
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalJobDatabase()

	var i job.Database
	i = d

	i.Print()
}
