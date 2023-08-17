package http_processor

import (
	"errors"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/utils"
	"net/http"
	"sync"
)

type Thread struct {
	mutex sync.Mutex

	Interrupt chan threads.InterruptEvent
	C12       chan common.ProcessorRequest
	C13       chan common.ProcessorResponse

	ProcessorResponseTable *utils.ResponseTable

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

	thread.C12, ok = (channels[1]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	thread.C13, ok = (channels[2]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	thread.ProcessorResponseTable = utils.NewResponseTable()

	return thread, nil
}
