package processor

import (
	"errors"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl-light/utils"
	"github.com/GabeCordo/etl/core/threads/common"
	"sync"
)

type Thread struct {
	Interrupt chan threads.InterruptEvent
	C5        chan common.ProcessorRequest
	C6        chan common.ProcessorResponse
	C12       chan common.ProcessorRequest
	C13       chan common.ProcessorResponse
	C14       chan common.SupervisorRequest
	C15       chan common.SupervisorResponse

	SupervisorResponseTable *utils.ResponseTable

	Logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewThread(logger *utils.Logger, channels ...any) (*Thread, error) {
	processor := new(Thread)

	if logger != nil {
		processor.Logger = logger
	} else {
		return nil, errors.New("logger cannot be nil")
	}

	var ok bool = false

	processor.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	processor.C5, ok = (channels[1]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	processor.C6, ok = (channels[2]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	processor.C12, ok = (channels[3]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}

	processor.C13, ok = (channels[4]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}

	processor.C14, ok = (channels[5]).(chan common.SupervisorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 5")
	}

	processor.C15, ok = (channels[6]).(chan common.SupervisorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 6")
	}

	return processor, nil
}
