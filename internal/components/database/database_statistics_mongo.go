package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStatisticsDatabase struct {
	client *mongo.Client
}

func NewMongoStatisticsDatabase(uri string) (*MongoStatisticsDatabase, error) {
	database := new(MongoStatisticsDatabase)

	var err error
	database.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (database *MongoStatisticsDatabase) Get(filter StatisticFilter) (records []Statistic, err error) {

	d := database.client.Database("cluster-tools")
	c := d.Collection("statistics")

	var mongoFilter bson.D
	if filter.Cluster == "" {
		mongoFilter = bson.D{{"module", filter.Module}}
	} else {
		mongoFilter = bson.D{
			{"$and",
				bson.A{
					bson.D{{"module", bson.D{{"$eq", filter.Module}}}},
					bson.D{{"cluster", bson.D{{"$eq", filter.Cluster}}}},
				},
			},
		}
	}

	cursor, err := c.Find(context.TODO(), mongoFilter)
	if err != nil {
		return nil, err
	}

	err = cursor.Decode(records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (database *MongoStatisticsDatabase) Create(moduleId, clusterId string, statistic Statistic) (err error) {

	d := database.client.Database("cluster-tools")
	c := d.Collection("statistics")

	_, err = c.InsertOne(context.TODO(), statistic)
	if err != nil {
		return err
	}

	return nil
}

func (database *MongoStatisticsDatabase) Delete(moduleId string) (err error) {

	d := database.client.Database("cluster-tools")
	c := d.Collection("statistics")

	mongoFilter := bson.D{{"module", moduleId}}
	_, err = c.DeleteMany(context.TODO(), mongoFilter)
	if err != nil {
		return err
	}

	return nil
}
