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

const DatabaseName string = "flock"
const CollectionName string = "statistics"

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

	d := db.client.Database(DatabaseName)
	c := d.Collection(CollectionName)

	var mongoFilter bson.D
	if filter.Namespace == "" {
		mongoFilter = bson.D{{"namespace", filter.Namespace}}
	} else {
		mongoFilter = bson.D{
			{"$and",
				bson.A{
					bson.D{{"namespace", bson.D{{"$eq", filter.Namespace}}}},
					bson.D{{"pipeline", bson.D{{"$eq", filter.Pipeline}}}},
				},
			},
		}
	}

	cursor, err := c.Find(context.TODO(), mongoFilter)
	if err != nil {
		return records
	}

	stats := make([]*Wrapper, 0)
	err = cursor.All(context.TODO(), &stats)
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

	d := db.client.Database(DatabaseName)
	c := d.Collection(CollectionName)

	_, err = c.InsertOne(context.TODO(), statistic)
	return result, err
}

func (db *MongoStatisticsDatabase) Delete(filter database.Filter) (err error) {

	d := db.client.Database(DatabaseName)
	c := d.Collection(CollectionName)

	mongoFilter := bson.D{{"namespace", filter.Namespace}}
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
