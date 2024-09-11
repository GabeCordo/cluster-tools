package cache

import (
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
)

// processSaveRequest
// will insert or override an existing cache record
func (thread *Thread) processSaveRequest(request *common.ThreadRequest) {

	response := common.ThreadResponse{Source: common.Cache, Nonce: request.Nonce}

	cacheRequestData, ok := (request.Data).(common.CacheRequestData)
	if !ok {
		response.Success = false
		thread.C10 <- response
		return
	}
	response.Data = common.CacheResponseData{Identifier: cacheRequestData.Identifier}

	if _, found := GetCacheInstance().Get(cacheRequestData.Identifier); found {
		response.Success = GetCacheInstance().Swap(cacheRequestData.Identifier, cacheRequestData.Data, cacheRequestData.ExpiresIn)
	} else {
		// what if the user forgets to pass in an expiry time that's now set to 0?
		var newIdentifier string
		if cacheRequestData.ExpiresIn == 0 {
			newIdentifier = GetCacheInstance().Save(cacheRequestData.Data)
		} else {
			newIdentifier = GetCacheInstance().Save(cacheRequestData.Data, cacheRequestData.ExpiresIn)
		}
		response.Success = true
		response.Data = common.CacheResponseData{Identifier: newIdentifier}
	}
	thread.C10 <- response
}

func (thread *Thread) processLoadRequest(request *common.ThreadRequest) {

	cacheRequestData, ok := (request.Data).(common.CacheRequestData)
	if !ok {
		return
	}

	cacheData, isFoundAndNotExpired := GetCacheInstance().Get(cacheRequestData.Identifier)
	thread.C10 <- common.ThreadResponse{
		Data: common.CacheResponseData{
			Identifier: cacheRequestData.Identifier,
			Data:       cacheData,
		},
		Success: isFoundAndNotExpired && (cacheData != nil),
		Nonce:   request.Nonce,
	}
}

func (thread *Thread) processPingCache(request *common.ThreadRequest) {

	if thread.config.Debug {
		thread.logger.Println("received ping over C9")
	}

	thread.C10 <- common.ThreadResponse{Nonce: request.Nonce, Success: true}
}
