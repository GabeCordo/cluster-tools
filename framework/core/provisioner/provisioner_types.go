package provisioner

import (
	"errors"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/utils"
	"sync"
)

type Thread struct {
	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C5 chan threads.ProvisionerRequest    // Supervisor is receiving core from the http_thread
	C6 chan<- threads.ProvisionerResponse // Supervisor is sending responses to the http_thread

	C7 chan<- threads.DatabaseRequest  // Supervisor is sending core to the database
	C8 <-chan threads.DatabaseResponse // Supervisor is receiving responses from the database

	C9  chan<- threads.CacheRequest  // Provisioner is sending requests to the cache
	C10 <-chan threads.CacheResponse // Provisioner is receiving responses from the CacheThread

	C11 chan<- threads.MessengerRequest // Provisioner is sending request to the messenger

	databaseResponseTable *utils.ResponseTable
	cacheResponseTable    *utils.ResponseTable

	logger *utils.Logger

	accepting   bool
	listenersWg sync.WaitGroup
	requestWg   sync.WaitGroup
}

func NewThread(logger *utils.Logger, channels ...interface{}) (*Thread, error) {
	provisioner := new(Thread)
	var ok bool

	provisioner.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	provisioner.C5, ok = (channels[1]).(chan threads.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	provisioner.C6, ok = (channels[2]).(chan threads.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}
	provisioner.C7, ok = (channels[3]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	provisioner.C8, ok = (channels[4]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	provisioner.C9, ok = (channels[5]).(chan threads.CacheRequest)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 5")
	}
	provisioner.C10, ok = (channels[6]).(chan threads.CacheResponse)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 6")
	}
	provisioner.C11, ok = (channels[7]).(chan threads.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}

	provisioner.databaseResponseTable = utils.NewResponseTable()
	provisioner.cacheResponseTable = utils.NewResponseTable()

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	provisioner.logger = logger
	provisioner.logger.SetColour(utils.Orange)

	return provisioner, nil
}
