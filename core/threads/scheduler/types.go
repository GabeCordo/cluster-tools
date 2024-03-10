package scheduler

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/components/scheduler"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
)

type Config struct {
	Debug   bool
	Timeout float64
}

type Thread struct {
	Interrupt chan<- common.InterruptEvent

	C18 chan<- common.ProcessorRequest  // Processor rec req from the http_client thread
	C19 <-chan common.ProcessorResponse // Processor sending rsp to the http_client thread

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

	thread.C18, ok = (channels[1]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C19, ok = (channels[2]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.responseTable = multithreaded.NewResponseTable()

	return thread, nil
}
