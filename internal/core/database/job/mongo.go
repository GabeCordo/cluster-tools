package job

import (
	"context"
	"errors"
	"log"

	"github.com/FortifiedCode/flock/internal/core/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDatabase struct {
	client *mongo.Client
}

func NewMongoJobDatabase(uri string) (*MongoDatabase, error) {

	db := new(MongoDatabase)

	var err error

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	db.client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db MongoDatabase) Get(filter database.Filter) (records []any) {

	d := db.client.Database("flock")
	c := d.Collection("jobs")

	var mongoFilter bson.D
	if filter.UseNamespace() {
		mongoFilter = bson.D{{"module", bson.D{{"$eq", filter.Namespace}}}}
	} else if filter.UsePipeline() {
		mongoFilter = bson.D{
			{"$and",
				bson.A{
					bson.D{{"module", bson.D{{"$eq", filter.Namespace}}}},
					bson.D{{"cluster", bson.D{{"$eq", filter.Pipeline}}}},
				},
			},
		}
	} else if filter.UseIdentifier() {
		mongoFilter = bson.D{{"identifier", bson.D{{"$eq", filter.Identifier}}}}
	} else {
		mongoFilter = bson.D{}
	}

	cursor, err := c.Find(context.TODO(), mongoFilter)
	if err != nil {
		return records
	}

	jobs := make([]*Job, 0)
	err = cursor.All(context.TODO(), &jobs)

	records = make([]any, len(jobs))
	for i, job := range jobs {
		records[i] = job
	}

	return records
}

func (db MongoDatabase) Create(filter database.Filter, record any) (result any, err error) {

	job, ok := (record).(*Job)
	if !ok || (job == nil) {
		return nil, errors.New("job can not be nil")
	}

	d := db.client.Database("flock")
	c := d.Collection("jobs")

	_, err = c.InsertOne(context.TODO(), job)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (db MongoDatabase) Delete(filter database.Filter) (err error) {

	d := db.client.Database("flock")
	c := d.Collection("jobs")

	var mongoFilter bson.D
	if filter.UseNamespace() {
		mongoFilter = bson.D{{"module", bson.D{{"$eq", filter.Namespace}}}}
	} else if filter.UsePipeline() {
		mongoFilter = bson.D{
			{"$and",
				bson.A{
					bson.D{{"module", bson.D{{"$eq", filter.Namespace}}}},
					bson.D{{"cluster", bson.D{{"$eq", filter.Pipeline}}}},
				},
			},
		}
	} else if filter.UseIdentifier() {
		mongoFilter = bson.D{{"identifier", bson.D{{"$eq", filter.Identifier}}}}
	}

	result, err := c.DeleteOne(context.TODO(), mongoFilter)
	if err != nil {
		return err
	}

	if result.DeletedCount < 1 {
		return errors.New("no jobs were deleted")
	}

	return nil
}

func (db MongoDatabase) Replace(filter database.Filter, record any) (err error) {

	log.Println("not implemented")
	return err
}

func (db MongoDatabase) Save(path string) (err error) {

	log.Println("not implemented")
	return err
}

func (db MongoDatabase) Load(path string) (err error) {

	log.Println("not implemented")
	return err
}

func (db MongoDatabase) Print() {

	log.Println("not implemented")
}
