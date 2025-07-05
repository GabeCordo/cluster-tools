package processor

import (
	"errors"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/Flock/internal/shared/nonce"
	"github.com/GabeCordo/toolchain/logging"
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
	channels struct {
		interrupt <-chan thread.InterruptEvent

		c5 <-chan *thread.Request  // Processor rec req from the rest thread
		c6 chan<- *thread.Response // Processor sending rsp to the rest thread

		c7 <-chan *thread.Request  // Processor rec req from the processor thread
		c8 chan<- *thread.Response // Processor sending rsp to the processor thread

		c11 chan<- *thread.Request  // Processor sending req to the database thread
		c12 <-chan *thread.Response // Processor rec rsp from the database thread

		c13 chan<- *thread.Request  // Processor thread sending req to the runner thread
		c14 <-chan *thread.Response // Processor thread rec rsp from the runner thread

		c18 <-chan *thread.Request  // Processor rec req from the scheduler thread
		c19 chan<- *thread.Response // Processor sending rsp to the scheduler thread

		close chan thread.InterruptEvent
	}

	requestStore map[nonce.Nonce]*thread.Request

	processorTable *processor.Table

	config *Config
	Logger *logging.Logger

	wg sync.WaitGroup
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

	var ok = false

	t.channels.interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	t.channels.c5, ok = (channels[1]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.channels.c6, ok = (channels[2]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	t.channels.c7, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}

	t.channels.c8, ok = (channels[4]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}

	t.channels.c11, ok = (channels[5]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}

	t.channels.c12, ok = (channels[6]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}

	t.channels.c13, ok = (channels[7]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 7")
	}

	t.channels.c14, ok = (channels[8]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 8")
	}

	t.channels.c18, ok = (channels[9]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 9")
	}

	t.channels.c19, ok = (channels[10]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 10")
	}

	t.channels.close = make(chan thread.InterruptEvent)

	t.requestStore = make(map[nonce.Nonce]*thread.Request)

	return t, nil
}
