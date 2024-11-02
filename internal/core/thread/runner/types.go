package runner

import (
	"errors"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"github.com/Sentmint/PipelineOps/internal/core/database"
	"github.com/Sentmint/PipelineOps/internal/core/thread"
	"sync"
)

type Config struct {
	Debug   bool
	Timeout float64
}

type Thread struct {
	Interrupt chan thread.InterruptEvent

	channels struct {
		C13 chan thread.Request  // runner receives requests from the processor
		C14 chan thread.Response // runner sends responses to the processor

		C15 chan thread.Request  //runner sends requests to the database
		C16 chan thread.Response // runner receives responses from the database

		C9  chan thread.Request  // runner sends requests to the tls-socket
		C10 chan thread.Response // runner receives responses from the tls-socket

		C17 chan thread.Request // runner sends requests to the messenger
	}

	responseTable struct {
		database *multithreaded.ResponseTable
		socket   *multithreaded.ResponseTable
	}

	config *Config
	Logger *logging.Logger

	registry database.Database

	accepting bool
	wg        sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, registry database.Database, channels ...any) (*Thread, error) {
	t := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	if logger != nil {
		t.Logger = logger
	} else {
		return nil, errors.New("expected logger to be a non-nil value")
	}

	var ok bool = false

	t.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	t.channels.C13, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 1")
	}
	t.channels.C14, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 2")
	}
	t.channels.C15, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	t.channels.C16, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	t.channels.C17, ok = (channels[5]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 5")
	}
	t.channels.C9, ok = (channels[6]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan RunnerRequest' in index 6")
	}
	t.channels.C10, ok = (channels[7]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan RunnerResponse' in index 7")
	}

	t.responseTable.database = multithreaded.NewResponseTable()
	t.responseTable.socket = multithreaded.NewResponseTable()

	t.registry = registry

	return t, nil
}
