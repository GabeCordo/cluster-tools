package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Driver struct {
	client *mongo.Client
	meta   struct {
		uri string
	}
	flags struct {
		connected bool
	}
}

func NewDriver(uri string) Driver {

	driver := Driver{}
	driver.flags.connected = false
	driver.meta.uri = uri
	return driver
}

func (driver *Driver) IsConnected() bool {

	return driver.flags.connected
}

func (driver *Driver) Connect() (err error) {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(driver.meta.uri).SetServerAPIOptions(serverAPI)

	driver.client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		driver.flags.connected = true
	}

	return err
}

func (driver *Driver) Database(name string) Database {

	d := driver.client.Database(name)
	return Database{d}
}
