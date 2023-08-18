package processor

import (
	"errors"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl-light/utils"
	"github.com/GabeCordo/etl/core/threads/common"
	"sync"
)

type Thread struct {
	Interrupt chan<- threads.InterruptEvent

	C5 <-chan common.ProcessorRequest  // Processor rec req from the http_client thread
	C6 chan<- common.ProcessorResponse // Processor sending rsp to the http_client thread

	C7 <-chan common.ProcessorRequest  // Processor rec req from the http_processor thread
	C8 chan<- common.ProcessorResponse // Processor sending rsp to the http_processor thread

	C11 chan<- threads.DatabaseRequest  // Processor sending req to the database thread
	C12 <-chan threads.DatabaseResponse // Processor rec rsp from the database thread

	C13 chan<- common.SupervisorRequest  // Processor thread sending req to the supervisor thread
	C14 <-chan common.SupervisorResponse // Processor thread rec rsp from the supervisor thread

	SupervisorResponseTable *utils.ResponseTable
	DatabaseResponseTable   *utils.ResponseTable

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

	processor.C7, ok = (channels[3]).(chan common.ProcessorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}

	processor.C8, ok = (channels[4]).(chan common.ProcessorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}

	processor.C11, ok = (channels[5]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}

	processor.C12, ok = (channels[6]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}

	processor.C13, ok = (channels[7]).(chan common.SupervisorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 7")
	}

	processor.C14, ok = (channels[8]).(chan common.SupervisorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 8")
	}

	processor.SupervisorResponseTable = utils.NewResponseTable()
	processor.DatabaseResponseTable = utils.NewResponseTable()

	return processor, nil
}
