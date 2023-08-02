package core

import (
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/components/cache"
	"time"
)

var CacheInstance *cache.Cache

func GetCacheInstance() *cache.Cache {
	if CacheInstance == nil {
		CacheInstance = cache.NewCache()
	}
	return CacheInstance
}

func (cacheThread *CacheThread) Setup() {
	cacheThread.accepting = true
}

func (cacheThread *CacheThread) Start() {
	go func() {
		// request from http_server
		for request := range cacheThread.C9 {
			if !cacheThread.accepting {
				break
			}
			cacheThread.wg.Add(1)
			cacheThread.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// cleaning the cacheThread of expired records
		for cacheThread.accepting {
			time.Sleep(1 * time.Minute)
			// every minute, attempt to clean the cacheThread by removing any records that
			// may have expired since we last checked
			GetCacheInstance().Clean()
		}
	}()

	cacheThread.wg.Wait()
}

func (cacheThread *CacheThread) Send(response *threads.CacheResponse) {

	cacheThread.C10 <- *response
}

func (cacheThread *CacheThread) ProcessIncomingRequest(request *threads.CacheRequest) {
	if request.Action == threads.CacheSaveIn {
		cacheThread.ProcessSaveRequest(request)
	} else if request.Action == threads.CacheLoadFrom {
		cacheThread.ProcessLoadRequest(request)
	} else if request.Action == threads.CacheLowerPing {
		cacheThread.ProcessPingCache(request)
	}

	cacheThread.wg.Done()
}

// ProcessSaveRequest will insert or override an existing cache record
func (cacheThread *CacheThread) ProcessSaveRequest(request *threads.CacheRequest) {
	var response threads.CacheResponse
	if _, found := GetCacheInstance().Get(request.Identifier); found {
		GetCacheInstance().Swap(request.Identifier, request.Data, request.ExpiresIn)
		response = threads.CacheResponse{Identifier: request.Identifier, Data: nil, Nonce: request.Nonce, Success: true}
	} else {
		newIdentifier := GetCacheInstance().Save(request.Data, request.ExpiresIn)
		response = threads.CacheResponse{Identifier: newIdentifier, Data: nil, Nonce: request.Nonce, Success: true}
	}
	cacheThread.C10 <- response
}

func (cacheThread *CacheThread) ProcessLoadRequest(request *threads.CacheRequest) {
	cacheData, isFoundAndNotExpired := GetCacheInstance().Get(request.Identifier)
	cacheThread.C10 <- threads.CacheResponse{
		Identifier: request.Identifier,
		Data:       cacheData,
		Nonce:      request.Nonce,
		Success:    isFoundAndNotExpired && (cacheData != nil),
	}
}

func (cacheThread *CacheThread) ProcessPingCache(request *threads.CacheRequest) {

	if GetConfigInstance().Debug {
		cacheThread.logger.Println("received ping over C9")
	}

	cacheThread.C10 <- threads.CacheResponse{Nonce: request.Nonce, Success: true}
}

func (cacheThread *CacheThread) Teardown() {
	cacheThread.accepting = false
}
