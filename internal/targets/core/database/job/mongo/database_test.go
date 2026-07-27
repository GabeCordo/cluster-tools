package mongo

import (
	"testing"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/drivers/mongo"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/job"
)

func Test_StatisticMongoDatabase_Is_Implemented(t *testing.T) {

	m := mongo.NewDriver("")
	d, err := NewMongoDatabase(m)
	if err != nil {
		t.Fatal(err)
	}

	var i job.Database
	i = d

	i.Print()
}
