package mongo

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	database *mongo.Database
}

func (database Database) Collection(name string) Collection {
	c := database.database.Collection(name)
	return Collection{c}
}
