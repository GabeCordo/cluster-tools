package core

import (
	"sync"
)

type DatabaseAction uint8

const (
	Read   DatabaseAction = 0
	Write                 = 1
	Delete                = 2
)

type DatabaseRequest struct {
	action     DatabaseAction
	parameters []string
}

type DatabaseResponse struct {
	success bool
}

type Database struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 <-chan DatabaseRequest  // Database is receiving core from the http_thread
	C2 chan<- DatabaseResponse // Database is sending responses to the http_thread

	C3 chan<- MessengerRequest  // Database is sending core to the Messenger
	C4 <-chan MessengerResponse // Database is receiving responses from the Messenger

	C7 <-chan DatabaseRequest  // Database is receiving core from the Supervisor
	C8 chan<- DatabaseResponse // Database is sending responses to the Supervisor

	accepting bool
	wg        sync.WaitGroup
}

func NewDatabase(channels ...interface{}) (*Database, bool) {
	database := new(Database)
	var ok bool

	database.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	database.C1, ok = (channels[1]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	database.C2, ok = (channels[2]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}
	database.C3, ok = (channels[3]).(chan MessengerRequest)
	if !ok {
		return nil, ok
	}
	database.C4, ok = (channels[4]).(chan MessengerResponse)
	if !ok {
		return nil, ok
	}
	database.C7, ok = (channels[5]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	database.C8, ok = (channels[6]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}

	return database, ok
}
