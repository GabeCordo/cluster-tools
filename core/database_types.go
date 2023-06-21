package core

import (
	"errors"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type DatabaseThread struct {
	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 <-chan threads.DatabaseRequest  // Database is receiving core from the http_thread
	C2 chan<- threads.DatabaseResponse // Database is sending responses to the http_thread

	C3 chan<- threads.MessengerRequest  // Database is sending core to the Messenger
	C4 <-chan threads.MessengerResponse // Database is receiving responses from the Messenger

	C7 <-chan threads.DatabaseRequest  // Database is receiving core from the Supervisor
	C8 chan<- threads.DatabaseResponse // Database is sending responses to the Supervisor

	messengerResponseTable *utils.ResponseTable

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewDatabase(logger *utils.Logger, channels ...interface{}) (*DatabaseThread, error) {
	database := new(DatabaseThread)
	var ok bool

	database.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	database.C1, ok = (channels[1]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	database.C2, ok = (channels[2]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	database.C3, ok = (channels[3]).(chan threads.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	database.C4, ok = (channels[4]).(chan threads.MessengerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 4")
	}
	database.C7, ok = (channels[5]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}
	database.C8, ok = (channels[6]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}

	database.messengerResponseTable = utils.NewResponseTable()

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	database.logger = logger
	database.logger.SetColour(utils.Purple)

	return database, nil
}
