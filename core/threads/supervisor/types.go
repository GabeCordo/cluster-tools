package supervisor

import (
	"errors"
	"github.com/GabeCordo/mango/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"sync"
)

type Config struct {
	Debug   bool
	Timeout float64
}

type Thread struct {
	Interrupt chan common.InterruptEvent

	C13 chan common.SupervisorRequest
	C14 chan common.SupervisorResponse

	C15 chan common.DatabaseRequest
	C16 chan common.DatabaseResponse

	C17 chan common.MessengerRequest

	config *Config
	Logger *logging.Logger

	DatabaseResponseTable *multithreaded.ResponseTable

	accepting bool
	wg        sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {
	thread := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *Config type")
	}
	thread.config = cfg

	if logger != nil {
		thread.Logger = logger
	} else {
		return nil, errors.New("expected logger to be a non-nil value")
	}

	var ok bool = false

	thread.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	thread.C13, ok = (channels[1]).(chan common.SupervisorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 1")
	}
	thread.C14, ok = (channels[2]).(chan common.SupervisorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 2")
	}
	thread.C15, ok = (channels[3]).(chan common.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	thread.C16, ok = (channels[4]).(chan common.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	thread.C17, ok = (channels[5]).(chan common.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}

	thread.DatabaseResponseTable = multithreaded.NewResponseTable()

	return thread, nil
}
