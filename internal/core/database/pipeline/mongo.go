package pipeline

import (
	"context"
	"errors"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/plover"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfigDatabase struct {
	client *mongo.Client
}

func NewMongoConfigDatabase(uri string) (*MongoConfigDatabase, error) {
	db := new(MongoConfigDatabase)

	var err error

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	db.client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db MongoConfigDatabase) Get(filter database.Filter) (records []any) {

	records = make([]any, 0)

	d := db.client.Database("modules")
	c := d.Collection(filter.Namespace)

	// when the identifier is empty we want to return all the configs in the database
	var mongoFilter bson.D
	if filter.Identifier != "" {
		mongoFilter = bson.D{{"identifier", filter.Identifier}}
	} else {
		mongoFilter = bson.D{}
	}

	cursor, err := c.Find(context.TODO(), mongoFilter)
	if err != nil {
		return records
	}

	var stats []plover.PipelineIR
	err = cursor.All(context.TODO(), &stats)
	if err != nil {
		return records
	}

	for _, stat := range stats {
		records = append(records, stat)
	}

	return records
}

func (db MongoConfigDatabase) Create(filter database.Filter, record any) (r any, err error) {

	cfg, ok := (record).(*plover.PipelineIR)
	if !ok {
		err = errors.New("expected type *plover.PipelineIR")
		return nil, err
	}

	d := db.client.Database("modules")
	c := d.Collection(filter.Namespace)

	records := db.Get(filter)
	numOfRecords := len(records)
	if numOfRecords >= 1 {
		return nil, errors.New("pipeline with the same identifier already exists in the module")
	}

	_, err = c.InsertOne(context.TODO(), cfg)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (db MongoConfigDatabase) Replace(filter database.Filter, record any) (err error) {

	cfg, ok := (record).(*plover.PipelineIR)
	if !ok {
		err = errors.New("expected type *plover.PipelineIR")
		return err
	}

	d := db.client.Database("modules")
	c := d.Collection(filter.Namespace)

	mongoFilter := bson.D{{"identifier", bson.D{{"$eq", filter.Identifier}}}}
	result, err := c.ReplaceOne(context.TODO(), mongoFilter, cfg)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("no pipeline with the specified identifier exist for the module; nothing to replace")
	}

	return nil
}

func (db MongoConfigDatabase) Delete(filter database.Filter) (err error) {

	d := db.client.Database("modules")
	c := d.Collection(filter.Namespace)

	mongoFilter := bson.D{{"identifier", bson.D{{"$eq", filter.Identifier}}}}
	result, err := c.DeleteOne(context.TODO(), mongoFilter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no pipeline with the specified identifier exist for the module; nothing to delete")
	}

	return nil
}

func (db MongoConfigDatabase) Save(path string) (err error) {

	log.Println("not implemented")
	return err
}

func (db MongoConfigDatabase) Load(path string) (err error) {

	log.Println("not implemented")
	return err
}

func (db MongoConfigDatabase) Print() {

	log.Println("not implemented")
}
