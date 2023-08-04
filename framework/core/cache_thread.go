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

	for cacheThread.accepting {
		lastTimeInterval := time.Now()

		select {
		case request := <-cacheThread.C9:
			{
				cacheThread.wg.Add(1)
				cacheThread.ProcessIncomingRequest(&request)
			}
		default:
			{
				// every minute, attempt to clean the cacheThread by removing any records that
				// may have expired since we last checked
				if time.Now().Sub(lastTimeInterval).Minutes() >= 1 {
					GetCacheInstance().Clean()
				}
			}
		}

		time.Sleep(1 * time.Millisecond)
	}

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
