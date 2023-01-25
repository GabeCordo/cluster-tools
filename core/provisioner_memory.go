package core

import "sync"

type ProvisionerMemory struct {
	cacheResponses map[uint32]chan CacheResponse // uint32 => CacheResponse
	cacheMutex     sync.RWMutex
}

func NewProvisionerResponses() *ProvisionerMemory {
	provisionerResponses := new(ProvisionerMemory)
	provisionerResponses.cacheResponses = make(map[uint32]chan CacheResponse)
	return provisionerResponses
}

var provisionerMemory *ProvisionerMemory

func GetProvisionerMemoryInstance() *ProvisionerMemory {
	if provisionerMemory == nil {
		provisionerMemory = NewProvisionerResponses()
	}

	return provisionerMemory
}

func (memory ProvisionerMemory) CreateCacheResponseHook(nonce uint32) chan CacheResponse {
	memory.cacheMutex.Lock()
	defer memory.cacheMutex.Unlock()

	channel := make(chan CacheResponse)
	memory.cacheResponses[nonce] = channel

	return channel
}

func (memory ProvisionerMemory) LinkCacheResponse(nonce uint32, record CacheResponse) {
	memory.cacheMutex.RLock()
	defer memory.cacheMutex.RUnlock()

	if channel, found := memory.cacheResponses[nonce]; found {
		channel <- record
	}
}
