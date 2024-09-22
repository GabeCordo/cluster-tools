package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"sync"
)

type Config struct {
	Debug      bool
	Timeout    float64
	ProbeEvery uint32
	MaxRetry   uint32
	Net        struct {
		Host string
		Port int
	}
}

type Thread struct {
	Interrupt chan<- thread.InterruptEvent

	C5 <-chan thread.Request  // Processor rec req from the client thread
	C6 chan<- thread.Response // Processor sending rsp to the client thread

	C7 <-chan thread.Request  // Processor rec req from the processor thread
	C8 chan<- thread.Response // Processor sending rsp to the processor thread

	C11 chan<- thread.Request  // Processor sending req to the database thread
	C12 <-chan thread.Response // Processor rec rsp from the database thread

	C13 chan<- thread.Request  // Processor thread sending req to the supervisor thread
	C14 <-chan thread.Response // Processor thread rec rsp from the supervisor thread

	C18 <-chan thread.Request  // Processor rec req from the scheduler thread
	C19 chan<- thread.Response // Processor sending rsp to the scheduler thread

	SupervisorResponseTable *multithreaded.ResponseTable
	DatabaseResponseTable   *multithreaded.ResponseTable

	processorTable *processor.Table

	config *Config
	Logger *logging.Logger

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, table *processor.Table, channels ...any) (*Thread, error) {
	t := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	if logger != nil {
		t.Logger = logger
	} else {
		return nil, errors.New("logger cannot be nil")
	}

	t.processorTable = table

	var ok bool = false

	t.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	t.C5, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.C6, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	t.C7, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}

	t.C8, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}

	t.C11, ok = (channels[5]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}

	t.C12, ok = (channels[6]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}

	t.C13, ok = (channels[7]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 7")
	}

	t.C14, ok = (channels[8]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 8")
	}

	t.C18, ok = (channels[9]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 9")
	}

	t.C19, ok = (channels[10]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 10")
	}

	t.SupervisorResponseTable = multithreaded.NewResponseTable()
	t.DatabaseResponseTable = multithreaded.NewResponseTable()

	return t, nil
}
