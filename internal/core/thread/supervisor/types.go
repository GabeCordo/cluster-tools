package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"sync"
)

type Config struct {
	Debug   bool
	Timeout float64
}

type Thread struct {
	Interrupt chan thread.InterruptEvent

	C13 chan thread.Request  // supervisor receives requests from the processor
	C14 chan thread.Response // supervisor sends responses to the processor

	C15 chan thread.Request  //supervisor sends requests to the database
	C16 chan thread.Response // supervisor receives responses from the database

	C17 chan thread.Request // supervisor sends requests to the messenger

	config *Config
	Logger *logging.Logger

	DatabaseResponseTable *multithreaded.ResponseTable

	registry database.Database

	accepting bool
	wg        sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, registry database.Database, channels ...any) (*Thread, error) {
	t := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *config type")
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
	t.C13, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 1")
	}
	t.C14, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 2")
	}
	t.C15, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	t.C16, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	t.C17, ok = (channels[5]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}

	t.DatabaseResponseTable = multithreaded.NewResponseTable()

	t.registry = registry

	return t, nil
}
