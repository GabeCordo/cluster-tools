package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
	"sync"
)

type Config struct {
	Debug bool
	Net   struct {
		Host string
		Port int
	}
	Timeout float64
}

type Thread struct {
	mutex sync.Mutex

	Interrupt chan<- thread.InterruptEvent

	C7 chan<- thread.Request  // HTTP Processor is sending req to the processor_thread
	C8 <-chan thread.Response // HTTP Processor is rec rsp from the processor_thread

	C9  chan<- thread.Request  // HTTP Processor is sending req to the cache_thread
	C10 <-chan thread.Response // HTTP Processor is rec rsp from the cache_thread

	ProcessorResponseTable *multithreaded.ResponseTable
	CacheResponseTable     *multithreaded.ResponseTable

	server *http.Server
	mux    *http.ServeMux

	config *Config
	Logger *logging.Logger

	accepting bool
}

func New(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {
	t := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	if logger != nil {
		t.Logger = logger
	} else {
		return nil, errors.New("expected logger to be a non-nil value")
	}

	var ok bool = false

	t.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	t.C7, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.C8, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	t.C9, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.C10, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	t.ProcessorResponseTable = multithreaded.NewResponseTable()
	t.CacheResponseTable = multithreaded.NewResponseTable()

	return t, nil
}
