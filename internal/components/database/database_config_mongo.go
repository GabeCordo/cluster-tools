package database

import (
	"context"
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfigDatabase struct {
	client *mongo.Client
}

func NewMongoConfigDatabase(uri string) (*MongoConfigDatabase, error) {
	database := new(MongoConfigDatabase)

	var err error
	database.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (database MongoConfigDatabase) Get(filter ConfigFilter) (records []interfaces.Config, err error) {

	d := database.client.Database("modules")
	c := d.Collection(filter.Module)

	// when the identifier is empty we want to return all the configs in the database
	if filter.Identifier == "" {
		mongoFilter := bson.D{{}}

		cursor, err := c.Find(context.TODO(), mongoFilter)
		if err != nil {
			return nil, err
		}

		err = cursor.All(context.TODO(), &records)
		if err != nil {
			return nil, err
		}
	} else {
		mongoFilter := bson.D{{"identifier", bson.D{{"$eq", filter.Identifier}}}}

		config := &interfaces.Config{}
		err = c.FindOne(context.TODO(), mongoFilter).Decode(&config)
		if err != nil {
			return nil, err
		}

		records = append(records, *config)
	}

	return records, nil
}

func (database MongoConfigDatabase) Create(moduleIdentifier, configIdentifier string, cfg interfaces.Config) (err error) {

	d := database.client.Database("modules")
	c := d.Collection(moduleIdentifier)

	records, err := database.Get(ConfigFilter{Module: moduleIdentifier, Identifier: configIdentifier})
	if len(records) >= 1 {
		return errors.New("config with the same identifier already exists in the module")
	}

	_, err = c.InsertOne(context.TODO(), cfg)
	if err != nil {
		return err
	}

	return nil
}

func (database MongoConfigDatabase) Replace(moduleIdentifier, configIdentifier string, cfg interfaces.Config) (err error) {

	d := database.client.Database("modules")
	c := d.Collection(moduleIdentifier)

	mongoFilter := bson.D{{"identifier", bson.D{{"$eq", configIdentifier}}}}
	result, err := c.ReplaceOne(context.TODO(), mongoFilter, cfg)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("no config with the specified identifier exist for the module; nothing to replace")
	}

	return nil
}

func (database MongoConfigDatabase) Delete(moduleIdentifier, configIdentifier string) (err error) {

	d := database.client.Database("modules")
	c := d.Collection(moduleIdentifier)

	mongoFilter := bson.D{{"identifier", bson.D{{"$eq", configIdentifier}}}}
	result, err := c.DeleteOne(context.TODO(), mongoFilter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no config with the specified identifier exist for the module; nothing to delete")
	}

	return nil
}
