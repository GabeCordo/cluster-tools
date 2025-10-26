package pipeline

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/drivers/mongo"
	"github.com/FortifiedCode/plover"
)

const DatabaseName string = "flock"
const CollectionName string = "pipelines"

type MongoDatabase struct {
	driver mongo.Driver
}

func NewMongoDatabase(driver mongo.Driver) (*MongoDatabase, error) {

	mongoDatabase := new(MongoDatabase)
	mongoDatabase.driver = driver
	return mongoDatabase, nil
}

// Get returns the *plover.PipelineIR records associated with the filter.
func (mongoDatabase MongoDatabase) Get(filter database.Filter) (records []any) {

	if !mongoDatabase.driver.IsConnected() {
		return records
	}

	records = make([]any, 0)

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	stats := make([]Pipeline, 0)

	// when the identifier is empty we want to return all the configs in the database
	var err error
	if (filter.Namespace != "") && (filter.Identifier != "") {
		err = c.FindByIds("namespace", filter.Namespace, "identifier", filter.Identifier, &stats)
	} else if (filter.Namespace != "") && (filter.Identifier == "") {
		err = c.FindById("namespace", filter.Namespace, &stats)
	} else {
		err = c.FindAll(&stats)
	}

	if err != nil {
		return records
	}

	for _, stat := range stats {
		records = append(records, stat.Data)
	}

	return records
}

// Create stores a *plover.PipelineIR record inside the pipeline.MongoDatabase.
func (mongoDatabase MongoDatabase) Create(filter database.Filter, record any) (r any, err error) {

	if !mongoDatabase.driver.IsConnected() {
		return nil, database.NotConnected
	}

	if filter.Namespace == "" {
		err = errors.New("filter.Namespace is required")
		return r, err
	}

	if filter.Identifier == "" {
		err = errors.New("filter.Identifier is required")
		return r, err
	}

	pipelineRecord, ok := (record).(*plover.PipelineIR)
	if !ok {
		err = errors.New("expected type *plover.PipelineIR")
		return r, err
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	records := mongoDatabase.Get(filter)
	numOfRecords := len(records)
	if numOfRecords >= 1 {
		err = errors.New("pipeline with the same identifier already exists in the module")
		return r, err
	}

	pipeline := Pipeline{
		Namespace:  filter.Namespace,
		Identifier: filter.Identifier,
		Data:       pipelineRecord,
	}

	err = c.InsertOne(&pipeline)
	return nil, err
}

// Replace swaps a *plover.PipelineIR record with another one in the pipeline.MongoDatabase.
func (mongoDatabase MongoDatabase) Replace(filter database.Filter, record any) (err error) {

	if !mongoDatabase.driver.IsConnected() {
		return database.NotConnected
	}

	if filter.Namespace == "" {
		err = errors.New("filter.Namespace is required")
		return err
	}

	if filter.Identifier == "" {
		err = errors.New("filter.Identifier is required")
		return err
	}

	pipelineRecord, ok := (record).(*plover.PipelineIR)
	if !ok {
		err = errors.New("expected type *plover.PipelineIR")
		return err
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	pipeline := Pipeline{
		Namespace:  filter.Namespace,
		Identifier: filter.Identifier,
		Data:       pipelineRecord,
	}

	err = c.ReplaceByIds("namespace", filter.Namespace, "identifier", filter.Identifier, &pipeline)
	return err
}

// Delete removes a *plover.PipelineIR record inside the pipeline.MongoDatabase.
func (mongoDatabase MongoDatabase) Delete(filter database.Filter) (err error) {

	if !mongoDatabase.driver.IsConnected() {
		return database.NotConnected
	}

	if filter.Namespace == "" {
		err = errors.New("filter.Namespace is required")
		return err
	}

	if filter.Identifier == "" {
		err = errors.New("filter.Identifier is required")
		return err
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	err = c.DeleteByIds("namespace", filter.Namespace, "identifier", filter.Identifier)
	return err
}

// Save is not implemented for the pipeline.MongoDatabase struct.
func (mongoDatabase MongoDatabase) Save(path string) (err error) {

	err = nil
	return err
}

// Load is not implemented for the pipeline.MongoDatabase struct.
func (mongoDatabase MongoDatabase) Load(path string) (err error) {

	err = nil
	return err
}

// Print is not implemented for the pipeline.MongoDatabase struct.
func (mongoDatabase MongoDatabase) Print() {

	// nop
}
