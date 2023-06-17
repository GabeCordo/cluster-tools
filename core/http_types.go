package core

import (
	"context"
	"errors"
	"github.com/GabeCordo/etl/components/utils"
	"net/http"
	"sync"
)

// Frontend Thread

const (
	RefreshTime    = 1
	DefaultTimeout = 5
)

type HttpThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan<- DatabaseRequest  // Core is sending core to the Database
	C2 <-chan DatabaseResponse // Core is receiving responses from the Database

	C5 chan<- ProvisionerRequest  // Core is sending core to the Database
	C6 <-chan ProvisionerResponse // Core is receiving responses from the Database

	databaseResponses   map[uint32]DatabaseResponse
	supervisorResponses map[uint32]ProvisionerResponse

	provisionerResponseTable *utils.ResponseTable
	databaseResponseTable    *utils.ResponseTable

	server    *http.Server
	mux       *http.ServeMux
	cancelCtx context.CancelFunc

	logger *utils.Logger

	accepting bool
	counter   uint32
	mutex     sync.Mutex
	wg        sync.WaitGroup
}

func NewHttp(logger *utils.Logger, channels ...interface{}) (*HttpThread, error) {
	core := new(HttpThread)

	var ok bool

	core.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	core.C1, ok = (channels[1]).(chan DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	core.C2, ok = (channels[2]).(chan DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	core.C5, ok = (channels[3]).(chan ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 3")
	}
	core.C6, ok = (channels[4]).(chan ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 4")
	}

	core.databaseResponses = make(map[uint32]DatabaseResponse)
	core.supervisorResponses = make(map[uint32]ProvisionerResponse)

	core.provisionerResponseTable = utils.NewResponseTable()
	core.databaseResponseTable = utils.NewResponseTable()

	core.server = new(http.Server)

	core.accepting = true
	core.counter = 0

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	core.logger = logger
	core.logger.SetColour(utils.Green)

	return core, nil
}
