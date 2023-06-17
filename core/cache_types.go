package core

import (
	"errors"
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type CacheAction uint8

const (
	CacheSaveIn CacheAction = iota
	CacheLoadFrom
	CacheWipe
	CacheLowerPing
)

type CacheRequest struct {
	Action     CacheAction
	Identifier string
	Nonce      uint32
	Data       any
	ExpiresIn  float64 // duration in minutes
}

type CacheResponse struct {
	Identifier string
	Nonce      uint32
	Data       any
	Success    bool
}

type CacheThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C9  <-chan CacheRequest
	C10 chan<- CacheResponse

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewCacheThread(logger *utils.Logger, channels ...any) (*CacheThread, error) {
	cache := new(CacheThread)
	var ok bool

	cache.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	cache.C9, ok = (channels[1]).(chan CacheRequest)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 1")
	}
	cache.C10, ok = (channels[2]).(chan CacheResponse)
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
