package processor

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
	Interrupt chan<- common.InterruptEvent

	C5 <-chan common.ProcessorRequest  // Processor rec req from the http_client thread
	C6 chan<- common.ProcessorResponse // Processor sending rsp to the http_client thread

	C7 <-chan common.ProcessorRequest  // Processor rec req from the http_processor thread
	C8 chan<- common.ProcessorResponse // Processor sending rsp to the http_processor thread

	C11 chan<- common.DatabaseRequest  // Processor sending req to the database thread
	C12 <-chan common.DatabaseResponse // Processor rec rsp from the database thread

	C13 chan<- common.SupervisorRequest  // Processor thread sending req to the supervisor thread
	C14 <-chan common.SupervisorResponse // Processor thread rec rsp from the supervisor thread

	C18 <-chan common.ProcessorRequest  // Processor rec req from the scheduler thread
	C19 chan<- common.ProcessorResponse // Processor sending rsp to the scheduler thread

	SupervisorResponseTable *multithreaded.ResponseTable
	DatabaseResponseTable   *multithreaded.ResponseTable

	config *Config
	Logger *logging.Logger

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {
	thread := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *config type")
	}
	thread.config = cfg

	if logger != nil {
		thread.Logger = logger
	} else {
		return nil, errors.New("logger cannot be nil")
	}

	var ok bool = false

	thread.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	thread.C5, ok = (channels[1]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C6, ok = (channels[2]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.C7, ok = (channels[3]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}

	thread.C8, ok = (channels[4]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}

	thread.C11, ok = (channels[5]).(chan common.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}

	thread.C12, ok = (channels[6]).(chan common.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}

	thread.C13, ok = (channels[7]).(chan common.SupervisorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 7")
	}

	thread.C14, ok = (channels[8]).(chan common.SupervisorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 8")
	}

	thread.C18, ok = (channels[9]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 9")
	}

	thread.C19, ok = (channels[10]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 10")
	}

	thread.SupervisorResponseTable = multithreaded.NewResponseTable()
	thread.DatabaseResponseTable = multithreaded.NewResponseTable()

	return thread, nil
}
