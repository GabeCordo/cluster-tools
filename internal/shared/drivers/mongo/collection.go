package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var FailedToInsert = errors.New("driver failed to insert record")
var FailedToDelete = errors.New("driver failed to delete record")

type Collection struct {
	collection *mongo.Collection
}

func (collection Collection) find(filter bson.D, to any) (err error) {

	var cursor *mongo.Cursor
	cursor, err = collection.collection.Find(context.TODO(), filter)
	if err != nil {
		return err
	}

	err = cursor.All(context.TODO(), to)
	return err
}

func (collection Collection) filterByOne(idA, valueA string) bson.D {

	return bson.D{{idA, bson.D{{"$eq", valueA}}}}
}

func (collection Collection) filterByTwo(idA, valueA, idB, valueB string) bson.D {

	return bson.D{
		{"$and",
			bson.A{
				bson.D{{idA, bson.D{{"$eq", valueA}}}},
				bson.D{{idB, bson.D{{"$eq", valueB}}}},
			},
		},
	}
}

func (collection Collection) FindById(id, value string, out any) (err error) {

	filter := collection.filterByOne(id, value)
	err = collection.find(filter, out)
	return err
}

func (collection Collection) FindByIds(idA, valueA, idB, valueB string, out any) (err error) {

	filter := collection.filterByTwo(idA, valueA, idB, valueB)
	err = collection.find(filter, out)
	return err
}

func (collection Collection) FindAll(out any) (err error) {

	filter := bson.D{}
	err = collection.find(filter, out)
	return err
}

func (collection Collection) delete(filter bson.D) (err error) {

	var result *mongo.DeleteResult
	result, err = collection.collection.DeleteOne(context.TODO(), filter)
	if (err == nil) || (result == nil) {
		return err
	}

	if result.DeletedCount != 1 {
		err = FailedToDelete
	}
	return err
}

func (collection Collection) DeleteById(id, value string) (err error) {

	filter := collection.filterByOne(id, value)
	return collection.delete(filter)
}

func (collection Collection) DeleteByIds(idA, valueA, idB, valueB string) (err error) {

	filter := collection.filterByTwo(idA, valueA, idB, valueB)
	return collection.delete(filter)
}

func (collection Collection) DeleteManyById(id, value string) (err error) {

	filter := collection.filterByOne(id, value)
	_, err = collection.collection.DeleteMany(context.TODO(), filter)
	return err
}

func (collection Collection) InsertOne(record any) (err error) {

	_, err = collection.collection.InsertOne(context.TODO(), record)
	if err != nil {
		err = FailedToInsert
	}
	return err
}

func (collection Collection) ReplaceByIds(idA, valueA, idB, valueB string, in any) (err error) {

	filter := collection.filterByTwo(idA, valueA, idB, valueB)

	result, err := collection.collection.ReplaceOne(context.TODO(), filter, &in)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		err = errors.New("no pipeline with the specified identifier exist for the module; nothing to replace")
	}

	return err
}
