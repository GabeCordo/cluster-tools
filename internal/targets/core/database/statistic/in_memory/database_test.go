package in_memory

import (
	"testing"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/statistic"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalDatabase()

	var i statistic.Database
	i = d

	i.Print()
}
