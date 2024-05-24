package scheduler

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/components/scheduler"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
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

	C18 chan<- common.ThreadRequest  // Processor rec req from the processor thread
	C19 <-chan common.ThreadResponse // Processor sending rsp to the processor thread

	C20 <-chan common.ThreadRequest  // Processor receives request from http_client thread
	C21 chan<- common.ThreadResponse // Processor sends response to http_client thread

	wg sync.WaitGroup

	config *Config

	logger *logging.Logger

	responseTable *multithreaded.ResponseTable
	Scheduler     *scheduler.Scheduler
}

func New(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {

	thread := new(Thread)
	var ok = false

	if cfg == nil {
		panic("cfg passed to Scheduler thread must not be nil")
	}
	thread.config = cfg

	if logger == nil {
		panic("logger passed to Scheduler thread must not be nil")
	}
	thread.logger = logger

	thread.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	thread.C18, ok = (channels[1]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C19, ok = (channels[2]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.C20, ok = (channels[3]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SchedulerRequest' in index 3")
	}

	thread.C21, ok = (channels[4]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan SchedulerResponse' in index 4")
	}

	thread.responseTable = multithreaded.NewResponseTable()

	return thread, nil
}
