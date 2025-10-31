package mongo

import (
	"github.com/FortifiedCode/flock/internal/core/database/pipeline"
	"github.com/FortifiedCode/flock/internal/drivers/mongo"
	"testing"
)

func Test_StatisticMongoDatabase_Is_Implemented(t *testing.T) {

	m := mongo.NewDriver("")
	d, err := NewMongoDatabase(m)
	if err != nil {
		t.Fatal(err)
	}

	var i pipeline.Database
	i = d

	i.Print()
}
