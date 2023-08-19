package database

import (
	"errors"
	"github.com/GabeCordo/mango/threads"
	"github.com/GabeCordo/mango/utils"
	"sync"
)

type Thread struct {
	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 <-chan threads.DatabaseRequest  // Database is receiving threads from the http_thread
	C2 chan<- threads.DatabaseResponse // Database is sending responses to the http_thread

	C3 chan<- threads.MessengerRequest  // Database is sending threads to the Messenger
	C4 <-chan threads.MessengerResponse // Database is receiving responses from the Messenger

	C11 <-chan threads.DatabaseRequest  // Database is receiving req from the processor_thread
	C12 chan<- threads.DatabaseResponse // Database is sending rsp to the processor_thread

	C15 <-chan threads.DatabaseRequest  // Database is receiving req from the supervisor_thread
	C16 chan<- threads.DatabaseResponse // Database is sending rsp from the supervisor_thread

	messengerResponseTable *utils.ResponseTable

	configFolderPath    string
	statisticFolderPath string

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewThread(logger *utils.Logger, configPath, statisticPath string, channels ...interface{}) (*Thread, error) {
	database := new(Thread)
	var ok bool

	database.configFolderPath = configPath
	database.statisticFolderPath = statisticPath

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
	database.C11, ok = (channels[5]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}
	database.C12, ok = (channels[6]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}
	database.C15, ok = (channels[7]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 7")
	}
	database.C16, ok = (channels[8]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 8")
	}

	database.messengerResponseTable = utils.NewResponseTable()

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	database.logger = logger
	database.logger.SetColour(utils.Purple)

	return database, nil
}
