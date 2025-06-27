package scheduler

import (
	"errors"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/component/scheduler/job"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/thread"
	nonce2 "github.com/GabeCordo/Flock/internal/shared/nonce"
	"github.com/GabeCordo/toolchain/logging"
)

const nonceMin = 262114
const nonceMax = 524228 // (base) 262114 + 262114 (offset)

type Config struct {
	Debug           bool
	Timeout         float64
	SchedulesFolder string
}

type Thread struct {
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

	wg sync.WaitGroup

	config *Config

	logger *logging.Logger

	noncePool *nonce2.Pool

	processorResponseTable *nonce2.ResponseTable
	databaseResponseTable  *nonce2.ResponseTable

	jobDatabase database.Database

	Scheduler *job.Scheduler
}

func New(cfg *Config, logger *logging.Logger, jD database.Database, channels ...any) (*Thread, error) {

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

	t.noncePool = nonce2.New(nonceMin, nonceMax)

	t.processorResponseTable = nonce2.NewResponseTable()
	t.databaseResponseTable = nonce2.NewResponseTable()

	t.jobDatabase = jD

	return t, nil
}
