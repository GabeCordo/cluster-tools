package supervisor

import (
	"errors"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/utils"
	"sync"
)

type Thread struct {
	Interrupt chan threads.InterruptEvent

	C14 chan common.SupervisorRequest
	C15 chan common.SupervisorResponse

	C7 chan threads.DatabaseRequest
	C8 chan threads.DatabaseResponse

	C9  chan threads.CacheRequest
	C10 chan threads.CacheResponse

	C11 chan threads.MessengerRequest

	Logger *utils.Logger

	DatabaseResponseTable *utils.ResponseTable
	CacheResponseTable    *utils.ResponseTable

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
	thread.C14, ok = (channels[1]).(chan common.SupervisorRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 1")
	}
	thread.C15, ok = (channels[2]).(chan common.SupervisorResponse)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 2")
	}
	thread.C7, ok = (channels[3]).(chan threads.DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	thread.C8, ok = (channels[4]).(chan threads.DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	thread.C9, ok = (channels[5]).(chan threads.CacheRequest)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 5")
	}
	thread.C10, ok = (channels[6]).(chan threads.CacheResponse)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 6")
	}
	thread.C11, ok = (channels[7]).(chan threads.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}

	return thread, nil
}
