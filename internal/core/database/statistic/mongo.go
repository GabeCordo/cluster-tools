package statistic

import (
	"context"
	"errors"
	"github.com/FortifiedCode/flock/internal/core/database"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStatisticsDatabase struct {
	client *mongo.Client
}

func NewMongoStatisticsDatabase(uri string) (*MongoStatisticsDatabase, error) {
	db := new(MongoStatisticsDatabase)

	var err error

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	db.client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *MongoStatisticsDatabase) Get(filter database.Filter) (records []any) {

	d := db.client.Database("flock")
	c := d.Collection("statistics")

	var mongoFilter bson.D
	if filter.Identifier == "" {
		mongoFilter = bson.D{{"module", filter.Namespace}}
	} else {
		mongoFilter = bson.D{
			{"$and",
				bson.A{
					bson.D{{"module", bson.D{{"$eq", filter.Namespace}}}},
					bson.D{{"cluster", bson.D{{"$eq", filter.Identifier}}}},
				},
			},
		}
	}

	cursor, err := c.Find(context.TODO(), mongoFilter)
	if err != nil {
		return records
	}

	stats := make([]Wrapper, 0)
	err = cursor.Decode(stats)
	if err != nil {
		return records
	}

	records = make([]any, len(stats))
	for i, stat := range stats {
		records[i] = stat
	}

	return records
}

func (db *MongoStatisticsDatabase) Create(filter database.Filter, record any) (result any, err error) {

	result = nil

	statistic, ok := (record).(Wrapper)
	if !ok {
		err = errors.New("record is not a Wrapper")
		return result, err
	}

	d := db.client.Database("flock")
	c := d.Collection("statistics")

	_, err = c.InsertOne(context.TODO(), statistic)
	return result, err
}

func (db *MongoStatisticsDatabase) Delete(filter database.Filter) (err error) {

	d := db.client.Database("flock")
	c := d.Collection("statistics")

	mongoFilter := bson.D{{"module", filter.Namespace}}
	_, err = c.DeleteMany(context.TODO(), mongoFilter)
	if err != nil {
		return err
	}

	return nil
}

func (db *MongoStatisticsDatabase) Replace(filter database.Filter, record any) (err error) {

	log.Println("not implemented")
	return err
}

func (db *MongoStatisticsDatabase) Save(path string) (err error) {

	log.Println("not implemented")
	return err
}

func (db *MongoStatisticsDatabase) Load(path string) (err error) {

	log.Println("not implemented")
	return err
}

func (db *MongoStatisticsDatabase) Print() {

	log.Println("not implemented")
}
