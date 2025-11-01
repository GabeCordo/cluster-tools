package mongo

import (
	"github.com/FortifiedCode/flock/internal/shared/drivers/mongo"
	"github.com/FortifiedCode/flock/internal/targets/core/database"
	"github.com/FortifiedCode/flock/internal/targets/core/database/statistic"
	"github.com/FortifiedCode/plover"
	"time"
)

const DatabaseName string = "flock"
const CollectionName string = "statistics"

type MongoDatabase struct {
	driver mongo.Driver
}

func NewMongoDatabase(driver mongo.Driver) (*MongoDatabase, error) {

	mongoDatabase := new(MongoDatabase)
	mongoDatabase.driver = driver
	return mongoDatabase, nil
}

// Get retrieves *plover.Statistic records from the statistic.MongoDatabase.
func (mongoDatabase MongoDatabase) Get(filter database.Filter) (records []*plover.Statistics) {

	if !mongoDatabase.driver.IsConnected() {
		return records
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	stats := make([]*statistic.Statistic, 0)

	var err error
	if filter.Namespace == "" {
		err = c.FindById("namespace", filter.Namespace, &stats)
	} else {
		err = c.FindByIds("namespace", filter.Namespace, "pipeline", filter.Pipeline, &stats)
	}

	if err != nil {
		return records
	}

	records = make([]*plover.Statistics, len(stats))
	for i, stat := range stats {
		records[i] = stat.Data
	}

	return records
}

func (mongoDatabase MongoDatabase) Create(filter database.Filter, record *plover.Statistics) (result *plover.Statistics, err error) {

	result = nil
	if !mongoDatabase.driver.IsConnected() {
		return result, database.NotConnected
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	s := statistic.Statistic{
		Pipeline:  filter.Pipeline,
		Timestamp: time.Now(),
		Namespace: filter.Namespace,
		Data:      record,
	}

	err = c.InsertOne(&s)
	return result, err
}

func (mongoDatabase MongoDatabase) Delete(filter database.Filter) (err error) {

	if !mongoDatabase.driver.IsConnected() {
		return database.NotConnected
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	err = c.DeleteManyById("namespace", filter.Namespace)
	return err
}

// Replace is not implemented for the statistic.MongoDatabase.
func (mongoDatabase MongoDatabase) Replace(filter database.Filter, record *plover.Statistics) (err error) {

	err = database.NotImplemented
	return err
}

// Save is not implemented for the statistic.MongoDatabase.
func (mongoDatabase MongoDatabase) Save(path string) (err error) {

	err = database.NotImplemented
	return err
}

// Load is not implemented for the statistic.MongoDatabase.
func (mongoDatabase MongoDatabase) Load(path string) (err error) {

	err = database.NotImplemented
	return err
}

// Print is not implemented for the statistic.MongoDatabase.
func (mongoDatabase MongoDatabase) Print() {

	// nop
}
