package scheduler

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/scheduler/job"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"sync"
)

type Config struct {
	Debug           bool
	Timeout         float64
	SchedulesFolder string
}

type Thread struct {
	Interrupt chan<- thread.InterruptEvent

	C18 chan<- thread.Request  // Processor rec req from the processor thread
	C19 <-chan thread.Response // Processor sending rsp to the processor thread

	C20 <-chan thread.Request  // Processor receives request from client thread
	C21 chan<- thread.Response // Processor sends response to client thread

	C26 chan<- thread.Request  // Scheduler sends request to database_thread
	C27 <-chan thread.Response // Scheduler receives response from database_thread

	wg sync.WaitGroup

	config *Config

	logger *logging.Logger

	processorResponseTable *multithreaded.ResponseTable
	databaseResponseTable  *multithreaded.ResponseTable

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

	t.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	t.C18, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.C19, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	t.C20, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan SchedulerRequest' in index 3")
	}

	t.C21, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan SchedulerResponse' in index 4")
	}

	t.C26, ok = (channels[5]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}

	t.C27, ok = (channels[6]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}

	t.processorResponseTable = multithreaded.NewResponseTable()
	t.databaseResponseTable = multithreaded.NewResponseTable()

	t.jobDatabase = jD

	return t, nil
}
