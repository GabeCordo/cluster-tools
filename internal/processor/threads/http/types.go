package http

import (
	"context"
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
	"sync"
)

// Frontend Thread

type Config struct {
	Debug   bool
	Timeout float64
	Net     string
}

type Thread struct {
	Config *Config

	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan<- threads.ProvisionerRequest  // Core is sending threads to the Database
	C2 <-chan threads.ProvisionerResponse // Core is receiving responses from the Database

	ProvisionerResponseTable *multithreaded.ResponseTable

	server    *http.Server
	mux       *http.ServeMux
	cancelCtx context.CancelFunc

	logger *logging.Logger

	accepting bool
	counter   uint32
	mutex     sync.Mutex
	wg        sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, channels ...interface{}) (*Thread, error) {
	thread := new(Thread)

	var ok bool

	thread.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	thread.C1, ok = (channels[1]).(chan threads.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	thread.C2, ok = (channels[2]).(chan threads.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}

	thread.server = new(http.Server)

	thread.accepting = true
	thread.counter = 0

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	thread.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *http.pipeline type")
	}
	thread.Config = cfg

	thread.ProvisionerResponseTable = multithreaded.NewResponseTable()

	thread.logger.SetColour(logging.Green)

	return thread, nil
}
