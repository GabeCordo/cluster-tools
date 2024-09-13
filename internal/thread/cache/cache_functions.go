package cache

import (
	"github.com/GabeCordo/cluster-tools/internal/thread"
)

// processSaveRequest
// will insert or override an existing cache record
func (t *Thread) processSaveRequest(request *thread.Request) {

	response := thread.Response{Source: thread.Cache, Nonce: request.Nonce}

	cacheRequestData, ok := (request.Data).(thread.CacheRequestData)
	if !ok {
		response.Success = false
		t.C10 <- response
		return
	}
	response.Data = thread.CacheResponseData{Identifier: cacheRequestData.Identifier}

	if _, found := t.cache.Get(cacheRequestData.Identifier); found {
		response.Success = t.cache.Swap(cacheRequestData.Identifier, cacheRequestData.Data, cacheRequestData.ExpiresIn)
	} else {
		// what if the user forgets to pass in an expiry time that's now set to 0?
		var newIdentifier string
		if cacheRequestData.ExpiresIn == 0 {
			newIdentifier = t.cache.Save(cacheRequestData.Data)
		} else {
			newIdentifier = t.cache.Save(cacheRequestData.Data, cacheRequestData.ExpiresIn)
		}
		response.Success = true
		response.Data = thread.CacheResponseData{Identifier: newIdentifier}
	}
	t.C10 <- response
}

func (t *Thread) processLoadRequest(request *thread.Request) {

	cacheRequestData, ok := (request.Data).(thread.CacheRequestData)
	if !ok {
		return
	}

	cacheData, isFoundAndNotExpired := t.cache.Get(cacheRequestData.Identifier)
	t.C10 <- thread.Response{
		Data: thread.CacheResponseData{
			Identifier: cacheRequestData.Identifier,
			Data:       cacheData,
		},
		Success: isFoundAndNotExpired && (cacheData != nil),
		Nonce:   request.Nonce,
	}
}

func (t *Thread) processPingCache(request *thread.Request) {

	if t.config.Debug {
		t.logger.Println("received ping over C9")
	}

	t.C10 <- thread.Response{Nonce: request.Nonce, Success: true}
}
