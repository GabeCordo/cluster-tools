package cache

import (
	"errors"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/core/utils"
	"sync"
)

type Thread struct {
	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C9  <-chan threads.CacheRequest
	C10 chan<- threads.CacheResponse

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewThread(logger *utils.Logger, channels ...any) (*Thread, error) {
	cache := new(Thread)
	var ok bool

	cache.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	cache.C9, ok = (channels[1]).(chan threads.CacheRequest)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 1")
	}
	cache.C10, ok = (channels[2]).(chan threads.CacheResponse)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 2")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	cache.logger = logger
	cache.logger.SetColour(utils.Yellow)

	return cache, nil
}
