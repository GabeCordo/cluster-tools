package database

import (
	"context"
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JobMongoDatabase struct {
	client *mongo.Client
}

func NewMongoJobDatabase(uri string) (*JobMongoDatabase, error) {

	database := new(JobMongoDatabase)

	var err error
	database.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (database JobMongoDatabase) GetAll() (records []interfaces.Job, err error) {

	d := database.client.Database("cluster-tools")
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

func (database JobMongoDatabase) GetBy(filter *interfaces.Filter) (records []interfaces.Job, err error) {

	if filter == nil {
		return records, errors.New("filter can not be nil")
	}

	d := database.client.Database("cluster-tools")
	c := d.Collection("jobs")

	var mongoFilter bson.D
	if filter.UseModule() {
		mongoFilter = bson.D{{"module", bson.D{{"$eq", filter.Module}}}}
	} else if filter.UseCluster() {
		mongoFilter = bson.D{
			{"$and",
				bson.A{
					bson.D{{"module", bson.D{{"$eq", filter.Module}}}},
					bson.D{{"cluster", bson.D{{"$eq", filter.Cluster}}}},
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

func (database JobMongoDatabase) Create(job *interfaces.Job) (err error) {

	if job == nil {
		return errors.New("job can not be nil")
	}

	d := database.client.Database("cluster-tools")
	c := d.Collection("jobs")

	_, err = c.InsertOne(context.TODO(), job)
	if err != nil {
		return err
	}

	return nil
}

func (database JobMongoDatabase) Delete(filter *interfaces.Filter) (err error) {

	if filter == nil {
		return errors.New("filter can not be nil")
	}

	d := database.client.Database("cluster-tools")
	c := d.Collection("jobs")

	var mongoFilter bson.D
	if filter.UseModule() {
		mongoFilter = bson.D{{"module", bson.D{{"$eq", filter.Module}}}}
	} else if filter.UseCluster() {
		mongoFilter = bson.D{
			{"$and",
				bson.A{
					bson.D{{"module", bson.D{{"$eq", filter.Module}}}},
					bson.D{{"cluster", bson.D{{"$eq", filter.Cluster}}}},
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
