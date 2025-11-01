package mongo

import (
	"github.com/FortifiedCode/flock/internal/shared/drivers/mongo"
	"github.com/FortifiedCode/flock/internal/targets/core/database/statistic"
	"testing"
)

func Test_StatisticMongoDatabase_Is_Implemented(t *testing.T) {

	m := mongo.NewDriver("")
	d, err := NewMongoDatabase(m)
	if err != nil {
		t.Fatal(err)
	}

	var i statistic.Database
	i = d

	i.Print()
}
