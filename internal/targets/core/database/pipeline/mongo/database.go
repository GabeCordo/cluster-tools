package mongo

import (
	"errors"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/drivers/mongo"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/pipeline"
)

const DatabaseName string = "FunctionScheduler"
const CollectionName string = "pipelines"

type MongoDatabase struct {
	driver mongo.Driver
}

func NewMongoDatabase(driver mongo.Driver) (*MongoDatabase, error) {

	mongoDatabase := new(MongoDatabase)
	mongoDatabase.driver = driver
	return mongoDatabase, nil
}

// Get returns the *ScalingFunctions.PipelineIR records associated with the filter.
func (mongoDatabase MongoDatabase) Get(filter database.Filter) (records []*ScalingFunctions.PipelineIR) {

	if !mongoDatabase.driver.IsConnected() {
		return records
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	pp := make([]pipeline.Pipeline, 0)

	// when the identifier is empty we want to return all the configs in the database
	var err error
	if (filter.Namespace != "") && (filter.Identifier != "") {
		err = c.FindByIds("namespace", filter.Namespace, "identifier", filter.Identifier, &pp)
	} else if (filter.Namespace != "") && (filter.Identifier == "") {
		err = c.FindById("namespace", filter.Namespace, &pp)
	} else {
		err = c.FindAll(&pp)
	}

	if err != nil {
		return records
	}

	for _, p := range pp {
		records = append(records, p.Data)
	}

	return records
}

// Create stores a *ScalingFunctions.PipelineIR record inside the pipeline.MongoDatabase.
func (mongoDatabase MongoDatabase) Create(filter database.Filter, record *ScalingFunctions.PipelineIR) (id string, err error) {

	if !mongoDatabase.driver.IsConnected() {
		return id, database.NotConnected
	}

	if filter.Namespace == "" {
		err = errors.New("filter.Namespace is required")
		return id, err
	}

	if filter.Identifier == "" {
		err = errors.New("filter.Identifier is required")
		return id, err
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	records := mongoDatabase.Get(filter)
	numOfRecords := len(records)
	if numOfRecords >= 1 {
		err = errors.New("pipeline with the same identifier already exists in the module")
		return id, err
	}

	p := pipeline.Pipeline{
		Namespace:  filter.Namespace,
		Identifier: filter.Identifier,
		Data:       record,
	}
	err = c.InsertOne(&p)

	return id, err
}

// Replace swaps a *ScalingFunctions.PipelineIR record with another one in the pipeline.MongoDatabase.
func (mongoDatabase MongoDatabase) Replace(filter database.Filter, record *ScalingFunctions.PipelineIR) (err error) {

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

	p := pipeline.Pipeline{
		Namespace:  filter.Namespace,
		Identifier: filter.Identifier,
		Data:       record,
	}

	err = c.ReplaceByIds("namespace", filter.Namespace, "identifier", filter.Identifier, &p)
	return err
}

// Delete removes a *ScalingFunctions.PipelineIR record inside the pipeline.MongoDatabase.
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

func (mongoDatabase MongoDatabase) Distinct(filter database.Filter) (namespaces []any, err error) {

	if !mongoDatabase.driver.IsConnected() {
		err = database.NotConnected
		return namespaces, err
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	namespaces = c.Distinct("namespace")
	return namespaces, err
}
