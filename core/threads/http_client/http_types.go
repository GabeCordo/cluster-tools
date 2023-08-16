package http_client

import (
	"context"
	"errors"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/utils"
	"net/http"
	"sync"
)

// Frontend Thread

type Thread struct {
	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan<- threads.DatabaseRequest  // Core is sending threads to the Database
	C2 <-chan threads.DatabaseResponse // Core is receiving responses from the Database

	C5 chan<- common.ProcessorRequest  // Core is sending threads to the Database
	C6 <-chan common.ProcessorResponse // Core is receiving responses from the Database

	databaseResponses   map[uint32]threads.DatabaseResponse
	supervisorResponses map[uint32]threads.ProvisionerResponse

	ProcessorResponseTable *utils.ResponseTable
	DatabaseResponseTable  *utils.ResponseTable

	server    *http.Server
	mux       *http.ServeMux
	cancelCtx context.CancelFunc

	logger *utils.Logger

	accepting bool
	counter   uint32
	mutex     sync.Mutex
	wg        sync.WaitGroup
}

func NewThread(logger *utils.Logger, channels ...any) (*Thread, error) {
	core := new(Thread)

	var ok bool

	core.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	core.C1, ok = (channels[1]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	core.C2, ok = (channels[2]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	core.C5, ok = (channels[3]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}
	core.C6, ok = (channels[4]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}

	core.databaseResponses = make(map[uint32]threads.DatabaseResponse)
	core.supervisorResponses = make(map[uint32]threads.ProvisionerResponse)

	core.ProcessorResponseTable = utils.NewResponseTable()
	core.DatabaseResponseTable = utils.NewResponseTable()

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
