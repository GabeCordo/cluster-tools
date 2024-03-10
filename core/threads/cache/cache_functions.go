package cache

import (
	"github.com/GabeCordo/cluster-tools/core/threads/common"
)

// processSaveRequest
// will insert or override an existing cache record
func (thread *Thread) processSaveRequest(request *common.CacheRequest) {
	var response common.CacheResponse
	if _, found := GetCacheInstance().Get(request.Identifier); found {
		GetCacheInstance().Swap(request.Identifier, request.Data, request.ExpiresIn)
		response = common.CacheResponse{Identifier: request.Identifier, Data: nil, Nonce: request.Nonce, Success: true}
	} else {
		// what if the user forgets to pass in an expiry time that's now set to 0?
		var newIdentifier string
		if request.ExpiresIn == 0 {
			newIdentifier = GetCacheInstance().Save(request.Data)
		} else {
			newIdentifier = GetCacheInstance().Save(request.Data, request.ExpiresIn)
		}
		response = common.CacheResponse{Identifier: newIdentifier, Data: nil, Nonce: request.Nonce, Success: true}
	}
	thread.C10 <- response
}

func (thread *Thread) processLoadRequest(request *common.CacheRequest) {
	cacheData, isFoundAndNotExpired := GetCacheInstance().Get(request.Identifier)
	thread.C10 <- common.CacheResponse{
		Identifier: request.Identifier,
		Data:       cacheData,
		Nonce:      request.Nonce,
		Success:    isFoundAndNotExpired && (cacheData != nil),
	}
}

func (thread *Thread) processPingCache(request *common.CacheRequest) {

	if thread.config.Debug {
		thread.logger.Println("received ping over C9")
	}

	thread.C10 <- common.CacheResponse{Nonce: request.Nonce, Success: true}
}
