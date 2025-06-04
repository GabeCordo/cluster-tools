package job

import (
	"context"
	"errors"

	"github.com/GabeCordo/Flock/internal/core/database"
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
	db.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (database MongoDatabase) GetAll() (records []Job, err error) {

	d := database.client.Database("flock")
	c := d.Collection("jobs")

	cursor, err := c.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (database MongoDatabase) GetBy(filter *database.Filter) (records []Job, err error) {

	if filter == nil {
		return records, errors.New("filter can not be nil")
	}

	d := database.client.Database("flock")
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

	cursor, err := c.Find(context.TODO(), mongoFilter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (database MongoDatabase) Create(job *Job) (err error) {

	if job == nil {
		return errors.New("job can not be nil")
	}

	d := database.client.Database("flock")
	c := d.Collection("jobs")

	_, err = c.InsertOne(context.TODO(), job)
	if err != nil {
		return err
	}

	return nil
}

func (database MongoDatabase) Delete(filter *database.Filter) (err error) {

	if filter == nil {
		return errors.New("filter can not be nil")
	}

	d := database.client.Database("flock")
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
