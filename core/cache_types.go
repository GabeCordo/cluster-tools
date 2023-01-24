package core

import "sync"

type CacheAction uint8

const (
	SaveInCache CacheAction = iota
	LoadFromCache
	WipeCache
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

	accepting bool
	wg        sync.WaitGroup
}

func NewCacheThread(channels ...any) (*CacheThread, bool) {
	cache := new(CacheThread)
	var ok bool

	cache.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	cache.C9, ok = (channels[1]).(chan CacheRequest)
	if !ok {
		return nil, ok
	}
	cache.C10, ok = (channels[2]).(chan CacheResponse)
	if !ok {
		return nil, ok
	}

	return cache, ok
}
