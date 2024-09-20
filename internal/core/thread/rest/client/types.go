package client

import (
	"context"
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
	"sync"
)

// Frontend Thread

type Config struct {
	Debug      bool
	EnableCors bool
	Net        struct {
		Host string
		Port int
	}
	Timeout float64
}

type Thread struct {
	Interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan<- thread.Request  // Core is sending thread to the Database
	C2 <-chan thread.Response // Core is receiving responses from the Database

	C5 chan<- thread.Request  // Core is sending thread to the Database
	C6 <-chan thread.Response // Core is receiving responses from the Database

	C20 chan<- thread.Request  // ClientHttp is sending requests to the Scheduler
	C21 <-chan thread.Response // ClientHttp is receiving responses from the Scheduler

	C22 chan<- thread.Request  // Core is sending requests to the Messenger
	C23 <-chan thread.Response // Core is receiving responses from the Messenger

	C24 chan<- thread.Request  // Core is sending requests to the Cache
	C25 <-chan thread.Response // Core is receiving responses from the Cache

	ProcessorResponseTable *multithreaded.ResponseTable
	DatabaseResponseTable  *multithreaded.ResponseTable
	SchedulerResponseTable *multithreaded.ResponseTable
	MessengerResponseTable *multithreaded.ResponseTable
	CacheResponseTable     *multithreaded.ResponseTable

	server    *http.Server
	mux       *http.ServeMux
	cancelCtx context.CancelFunc

	config *Config
	logger *logging.Logger

	accepting bool
	counter   uint32
	mutex     sync.Mutex
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {
	t := new(Thread)

	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *config type")
	}
	t.config = cfg

	t.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	t.C1, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	t.C2, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	t.C5, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}
	t.C6, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}
	t.C20, ok = (channels[5]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 5")
	}
	t.C21, ok = (channels[6]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 6")
	}
	t.C22, ok = (channels[7]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}
	t.C23, ok = (channels[8]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 8")
	}
	t.C24, ok = (channels[9]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 9")
	}
	t.C25, ok = (channels[10]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 10")
	}

	t.ProcessorResponseTable = multithreaded.NewResponseTable()
	t.DatabaseResponseTable = multithreaded.NewResponseTable()
	t.SchedulerResponseTable = multithreaded.NewResponseTable()
	t.MessengerResponseTable = multithreaded.NewResponseTable()
	t.CacheResponseTable = multithreaded.NewResponseTable()

	t.server = new(http.Server)

	t.accepting = true
	t.counter = 0

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	t.logger = logger
	t.logger.SetColour(logging.Green)

	return t, nil
}
