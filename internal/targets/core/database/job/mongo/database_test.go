package mongo

import (
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/drivers/mongo"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
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
