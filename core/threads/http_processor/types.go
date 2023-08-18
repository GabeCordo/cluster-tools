package http_processor

import (
	"errors"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl-light/utils"
	"github.com/GabeCordo/etl/core/threads/common"
	"net/http"
	"sync"
)

type Thread struct {
	mutex sync.Mutex

	Interrupt chan<- threads.InterruptEvent

	C7 chan<- common.ProcessorRequest  // HTTP Processor is sending req to the processor_thread
	C8 <-chan common.ProcessorResponse // HTTP Processor is rec rsp from the processor_thread

	C9  chan<- threads.CacheRequest  // HTTP Processor is sending req to the cache_thread
	C10 <-chan threads.CacheResponse // HTTP Processor is rec rsp from the cache_thread

	ProcessorResponseTable *utils.ResponseTable
	CacheResponseTable     *utils.ResponseTable

	server *http.Server
	mux    *http.ServeMux

	Logger    *utils.Logger
	accepting bool
}

func NewThread(logger *utils.Logger, channels ...any) (*Thread, error) {
	thread := new(Thread)

	if logger != nil {
		thread.Logger = logger
	} else {
		return nil, errors.New("expected logger to be a non-nil value")
	}

	var ok bool = false

	thread.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	thread.C7, ok = (channels[1]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C8, ok = (channels[2]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.C9, ok = (channels[3]).(chan threads.CacheRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C10, ok = (channels[4]).(chan threads.CacheResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.ProcessorResponseTable = utils.NewResponseTable()
	thread.CacheResponseTable = utils.NewResponseTable()

	return thread, nil
}
