package cache

import (
	"github.com/GabeCordo/mango/core/threads/common"
	"time"
)

func (thread *Thread) Setup() {
	thread.accepting = true
}

func (thread *Thread) Start() {

	go func() {
		// request from http_server
		for request := range thread.C9 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.HttpProcessor
			thread.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// cleaning the thread of expired records
		for thread.accepting {
			time.Sleep(1 * time.Minute)
			// every minute, attempt to clean the thread by removing any records that
			// may have expired since we last checked
			GetCacheInstance().Clean()
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) Respond(response *common.CacheResponse) {

	thread.C10 <- *response
}

func (thread *Thread) ProcessIncomingRequest(request *common.CacheRequest) {
	if request.Action == common.CacheSaveIn {
		thread.ProcessSaveRequest(request)
	} else if request.Action == common.CacheLoadFrom {
		thread.ProcessLoadRequest(request)
	} else if request.Action == common.CacheLowerPing {
		thread.ProcessPingCache(request)
	}

	thread.wg.Done()
}

// ProcessSaveRequest will insert or override an existing cache record
func (thread *Thread) ProcessSaveRequest(request *common.CacheRequest) {
	var response common.CacheResponse
	if _, found := GetCacheInstance().Get(request.Identifier); found {
		GetCacheInstance().Swap(request.Identifier, request.Data, request.ExpiresIn)
		response = common.CacheResponse{Identifier: request.Identifier, Data: nil, Nonce: request.Nonce, Success: true}
	} else {
		newIdentifier := GetCacheInstance().Save(request.Data, request.ExpiresIn)
		response = common.CacheResponse{Identifier: newIdentifier, Data: nil, Nonce: request.Nonce, Success: true}
	}
	thread.C10 <- response
}

func (thread *Thread) ProcessLoadRequest(request *common.CacheRequest) {
	cacheData, isFoundAndNotExpired := GetCacheInstance().Get(request.Identifier)
	thread.C10 <- common.CacheResponse{
		Identifier: request.Identifier,
		Data:       cacheData,
		Nonce:      request.Nonce,
		Success:    isFoundAndNotExpired && (cacheData != nil),
	}
}

func (thread *Thread) ProcessPingCache(request *common.CacheRequest) {

	if thread.config.Debug {
		thread.logger.Println("received ping over C9")
	}

	thread.C10 <- common.CacheResponse{Nonce: request.Nonce, Success: true}
}

func (thread *Thread) Teardown() {
	thread.accepting = false
}
