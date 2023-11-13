package http_client

import (
	"context"
	"errors"
	"github.com/GabeCordo/mango/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
	"sync"
)

// Frontend Thread

type Config struct {
	Debug      bool
	EnableCors bool
	Net        struct {
		Host string
		Port int
	}
	Timeout float64
}

type Thread struct {
	Interrupt chan<- common.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan<- common.DatabaseRequest  // Core is sending threads to the Database
	C2 <-chan common.DatabaseResponse // Core is receiving responses from the Database

	C5 chan<- common.ProcessorRequest  // Core is sending threads to the Database
	C6 <-chan common.ProcessorResponse // Core is receiving responses from the Database

	databaseResponses   map[uint32]common.DatabaseResponse
	supervisorResponses map[uint32]common.SupervisorResponse

	ProcessorResponseTable *multithreaded.ResponseTable
	DatabaseResponseTable  *multithreaded.ResponseTable

	server    *http.Server
	mux       *http.ServeMux
	cancelCtx context.CancelFunc

	config *Config
	logger *logging.Logger

	accepting bool
	counter   uint32
	mutex     sync.Mutex
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {
	thread := new(Thread)

	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *config type")
	}
	thread.config = cfg

	thread.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	thread.C1, ok = (channels[1]).(chan common.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	thread.C2, ok = (channels[2]).(chan common.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	thread.C5, ok = (channels[3]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}
	thread.C6, ok = (channels[4]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}

	thread.databaseResponses = make(map[uint32]common.DatabaseResponse)
	thread.supervisorResponses = make(map[uint32]common.SupervisorResponse)

	thread.ProcessorResponseTable = multithreaded.NewResponseTable()
	thread.DatabaseResponseTable = multithreaded.NewResponseTable()

	thread.server = new(http.Server)

	thread.accepting = true
	thread.counter = 0

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	thread.logger = logger
	thread.logger.SetColour(logging.Green)

	return thread, nil
}
