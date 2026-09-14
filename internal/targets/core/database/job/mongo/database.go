package mongo

import (
	"github.com/GabeCordo/DistributedFunctions/internal/shared/drivers/mongo"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

const DatabaseName string = "DistributedFunctions"
const CollectionName string = "jobs"

type MongoDatabase struct {
	driver mongo.Driver
}

func NewMongoDatabase(driver mongo.Driver) (*MongoDatabase, error) {

	db := new(MongoDatabase)
	db.driver = driver
	return db, nil
}

// Get retrieves the *Job records from the job.MongoDatabase.
func (mongoDatabase MongoDatabase) Get(filter database.Filter) (jobs []*job.Job) {

	if !mongoDatabase.driver.IsConnected() {
		return jobs
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	var err error
	if filter.UseNamespace() {
		err = c.FindById("module", filter.Namespace, &jobs)
	} else if filter.UsePipeline() {
		err = c.FindByIds("module", filter.Namespace, "cluster", filter.Pipeline, &jobs)
	} else if filter.UseIdentifier() {
		err = c.FindById("identifier", filter.Identifier, &jobs)
	} else {
		err = c.FindAll(&jobs)
	}

	if err != nil {
		return jobs
	}

	return jobs
}

// Create adds a *Job record to the job.MongoDatabase.
func (mongoDatabase MongoDatabase) Create(filter database.Filter, record *job.Job) (id string, err error) {

	if !mongoDatabase.driver.IsConnected() {
		return "", database.NotConnected
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	err = c.InsertOne(record)
	return "", err
}

// Delete removes a *Job record from the job.MongoDatabase.
func (mongoDatabase MongoDatabase) Delete(filter database.Filter) (err error) {

	if !mongoDatabase.driver.IsConnected() {
		return database.NotConnected
	}

	d := mongoDatabase.driver.Database(DatabaseName)
	c := d.Collection(CollectionName)

	if filter.UseNamespace() {
		err = c.DeleteById("module", filter.Namespace)
	} else if filter.UsePipeline() {
		err = c.DeleteByIds("module", filter.Namespace, "cluster", filter.Pipeline)
	} else if filter.UseIdentifier() {
		err = c.DeleteById("identifier", filter.Identifier)
	}

	return err
}

// Replace swaps a *Job record from the job.MongoDatabase.
func (mongoDatabase MongoDatabase) Replace(filter database.Filter, record *job.Job) (err error) {

	err = database.NotImplemented
	return err
}

// Save is not implemented for the job.MongoDatabase struct.
func (mongoDatabase MongoDatabase) Save(path string) (err error) {

	err = database.NotImplemented
	return err
}

// Load is not implemented for the job.MongoDatabase struct.
func (mongoDatabase MongoDatabase) Load(path string) (err error) {

	err = database.NotImplemented
	return err
}

// Print is not implemented for the job.MongoDatabase struct.
func (mongoDatabase MongoDatabase) Print() {

	// nop
}
