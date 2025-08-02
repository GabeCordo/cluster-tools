package scheduler

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/Flock/internal/core/use_cases/scheduler"
	"github.com/GabeCordo/Flock/internal/shared/logging"
	nonce2 "github.com/GabeCordo/Flock/internal/shared/nonce"
)

type Config struct {
	Debug           bool
	Timeout         float64
	SchedulesFolder string
}

type Thread struct {
	config   *Config
	channels struct {
		interrupt chan<- thread.InterruptEvent

		c18 chan<- *thread.Request  // Processor rec req from the processor thread
		c19 <-chan *thread.Response // Processor sending rsp to the processor thread

		c20 <-chan *thread.Request  // Processor receives request from rest thread
		c21 chan<- *thread.Response // Processor sends response to rest thread

		c26 chan<- *thread.Request  // Scheduler sends request to database_thread
		c27 <-chan *thread.Response // Scheduler receives response from database_thread

		close chan thread.InterruptEvent
	}
	useCases               scheduler.UseCases
	logger                 logging.Logger
	noncePool              *nonce2.Pool
	processorResponseTable *nonce2.ResponseTable
	databaseResponseTable  *nonce2.ResponseTable
}

func New(cfg *Config, logger logging.Logger, useCases scheduler.UseCases, noncePool *nonce2.Pool, channels ...any) (*Thread, error) {

	t := new(Thread)
	var ok = false

	if cfg == nil {
		panic("cfg passed to Scheduler thread must not be nil")
	}
	t.config = cfg

	if logger == nil {
		panic("logger passed to Scheduler thread must not be nil")
	}
	t.logger = logger

	t.channels.interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	t.channels.c18, ok = (channels[1]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.channels.c19, ok = (channels[2]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	t.channels.c20, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan SchedulerRequest' in index 3")
	}

	t.channels.c21, ok = (channels[4]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan SchedulerResponse' in index 4")
	}

	t.channels.c26, ok = (channels[5]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}

	t.channels.c27, ok = (channels[6]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}

	t.channels.close = make(chan thread.InterruptEvent)

	t.noncePool = noncePool

	t.processorResponseTable = nonce2.NewResponseTable()
	t.databaseResponseTable = nonce2.NewResponseTable()

	t.useCases = useCases

	return t, nil
}
