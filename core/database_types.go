package core

import (
	"errors"
	"github.com/GabeCordo/etl/components/database"
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type DatabaseAction uint8

const (
	DatabaseStore DatabaseAction = iota
	DatabaseFetch
	DatabaseReplace
	DatabaseDelete
	DatabaseUpperPing
	DatabaseLowerPing
)

type DatabaseRequest struct {
	Action  DatabaseAction    `json:"Action"`
	Nonce   uint32            `json:"Nonce"`
	Origin  Module            `json:"origin"`
	Type    database.DataType `json:"type"`
	Cluster string            `json:"cluster"` // aka. Cluster Identifier
	Module  string            `json:"module"`  // aka. Module Identifier
	Data    any               `json:"data"`    // *cluster.Response `json:"Data"`
}

type DatabaseResponse struct {
	Nonce   uint32 `json:"Nonce"`
	Success bool   `json:"Success"`
	Data    any    `json:"statistics"` // []database.Entry or cluster.Config
}

type DatabaseThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 <-chan DatabaseRequest  // Database is receiving core from the http_thread
	C2 chan<- DatabaseResponse // Database is sending responses to the http_thread

	C3 chan<- MessengerRequest  // Database is sending core to the Messenger
	C4 <-chan MessengerResponse // Database is receiving responses from the Messenger

	C7 <-chan DatabaseRequest  // Database is receiving core from the Supervisor
	C8 chan<- DatabaseResponse // Database is sending responses to the Supervisor

	messengerResponseTable *utils.ResponseTable

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewDatabase(logger *utils.Logger, channels ...interface{}) (*DatabaseThread, error) {
	database := new(DatabaseThread)
	var ok bool

	database.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	database.C1, ok = (channels[1]).(chan DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	database.C2, ok = (channels[2]).(chan DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	database.C3, ok = (channels[3]).(chan MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	database.C4, ok = (channels[4]).(chan MessengerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 4")
	}
	database.C7, ok = (channels[5]).(chan DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}
	database.C8, ok = (channels[6]).(chan DatabaseResponse)
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
