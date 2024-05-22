package http_processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
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

	Interrupt chan<- common.InterruptEvent

	C7 chan<- common.ThreadRequest  // HTTP Processor is sending req to the processor_thread
	C8 <-chan common.ThreadResponse // HTTP Processor is rec rsp from the processor_thread

	C9  chan<- common.ThreadRequest  // HTTP Processor is sending req to the cache_thread
	C10 <-chan common.ThreadResponse // HTTP Processor is rec rsp from the cache_thread

	ProcessorResponseTable *multithreaded.ResponseTable
	CacheResponseTable     *multithreaded.ResponseTable

	server *http.Server
	mux    *http.ServeMux

	config *Config
	Logger *logging.Logger

	accepting bool
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
		return nil, errors.New("expected logger to be a non-nil value")
	}

	var ok bool = false

	thread.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	thread.C7, ok = (channels[1]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C8, ok = (channels[2]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.C9, ok = (channels[3]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C10, ok = (channels[4]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.ProcessorResponseTable = multithreaded.NewResponseTable()
	thread.CacheResponseTable = multithreaded.NewResponseTable()

	return thread, nil
}
