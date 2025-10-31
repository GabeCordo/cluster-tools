package in_memory

import (
	"github.com/FortifiedCode/flock/internal/core/database/statistic"
	"testing"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalDatabase()

	var i statistic.Database
	i = d

	i.Print()
}
