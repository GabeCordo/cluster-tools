package core

import "sync"

///////////////////////////////////////////////////////////////////////////
//							Cache Storage
//////////////////////////////////////////////////////////////////////////

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

func (memory *ProvisionerMemory) CreateCacheResponseEventListener(nonce uint32) chan CacheResponse {
	memory.cacheMutex.Lock()
	defer memory.cacheMutex.Unlock()

	channel := make(chan CacheResponse)
	memory.cacheResponses[nonce] = channel

	return channel
}

func (memory *ProvisionerMemory) SendCacheResponseEvent(nonce uint32, record CacheResponse) {
	memory.cacheMutex.RLock()
	defer memory.cacheMutex.RUnlock()

	if channel, found := memory.cacheResponses[nonce]; found {
		channel <- record
	}

	delete(memory.cacheResponses, nonce)
}

///////////////////////////////////////////////////////////////////////////
//							Cache Storage
//////////////////////////////////////////////////////////////////////////

type DatabaseMemory struct {
	databaseResponses map[uint32]chan DatabaseResponse // uint32 => CacheResponse
	cacheMutex        sync.RWMutex
}

func NewDatabaseResponses() *DatabaseMemory {
	databaseResponses := new(DatabaseMemory)
	databaseResponses.databaseResponses = make(map[uint32]chan DatabaseResponse)
	return databaseResponses
}

var databaseMemory *DatabaseMemory

func GetDatabaseMemoryInstance() *DatabaseMemory {
	if databaseMemory == nil {
		databaseMemory = NewDatabaseResponses()
	}
	return databaseMemory
}

func (memory *DatabaseMemory) CreateDatabaseResponseEventListener(nonce uint32) chan DatabaseResponse {
	memory.cacheMutex.Lock()
	defer memory.cacheMutex.Unlock()

	channel := make(chan DatabaseResponse)
	memory.databaseResponses[nonce] = channel

	return channel
}

func (memory *DatabaseMemory) SendDatabaseResponseEvent(nonce uint32, record DatabaseResponse) {
	memory.cacheMutex.RLock()
	defer memory.cacheMutex.RUnlock()

	if channel, found := memory.databaseResponses[nonce]; found {
		channel <- record
	}

	delete(memory.databaseResponses, nonce)
}
