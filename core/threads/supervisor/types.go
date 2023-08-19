package supervisor

import (
	"errors"
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango/threads"
	"github.com/GabeCordo/mango/utils"
	"sync"
)

type Thread struct {
	Interrupt chan threads.InterruptEvent

	C13 chan common.SupervisorRequest
	C14 chan common.SupervisorResponse

	C15 chan threads.DatabaseRequest
	C16 chan threads.DatabaseResponse

	C17 chan threads.MessengerRequest

	Logger *utils.Logger

	DatabaseResponseTable *utils.ResponseTable

	accepting bool
	wg        sync.WaitGroup
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
	thread.C13, ok = (channels[1]).(chan common.SupervisorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 1")
	}
	thread.C14, ok = (channels[2]).(chan common.SupervisorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 2")
	}
	thread.C15, ok = (channels[3]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	thread.C16, ok = (channels[4]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	thread.C17, ok = (channels[5]).(chan threads.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}

	thread.DatabaseResponseTable = utils.NewResponseTable()

	return thread, nil
}
