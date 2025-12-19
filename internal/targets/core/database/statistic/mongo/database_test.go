package mongo

import (
	"github.com/FortifiedCode/flock/internal/shared/drivers/mongo"
	"github.com/FortifiedCode/flock/internal/targets/core/database"
	"github.com/FortifiedCode/flock/internal/targets/core/database/statistic"
	"testing"
)

const TestDatabaseUri = "mongodb://localhost:27017"

func Test_StatisticMongoDatabase_Is_Implemented(t *testing.T) {

	m := mongo.NewDriver(TestDatabaseUri)
	d, err := NewMongoDatabase(m)
	if err != nil {
		t.Fatal(err)
	}

	var i statistic.Database
	i = d

	i.Print()
}

func Test_StatisticMongoDatabase_Summary(t *testing.T) {

	// LOCAL TEST
	// remove the .skip() directive when running the test locally.
	t.Skip()

	m := mongo.NewDriver(TestDatabaseUri)
	err := m.Connect()
	if err != nil {
		t.Fatal(err)
	}

	d, err := NewMongoDatabase(m)
	if err != nil {
		t.Fatal(err)
	}

	d.Distinct(database.Filter{})
}
