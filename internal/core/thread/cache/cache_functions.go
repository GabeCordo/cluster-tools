package cache

import (
	"github.com/GabeCordo/Flock/internal/core/thread"
)

// processSaveRequest
// will insert or override an existing cache record
func (t *Thread) processSaveRequest(request *thread.Request, response *thread.Response) {

	cacheRequestData, ok := (request.Data).(thread.CacheRequestData)
	if !ok {
		response.Success = false
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
}

func (t *Thread) processLoadRequest(request *thread.Request, response *thread.Response) {

	cacheRequestData, ok := (request.Data).(thread.CacheRequestData)
	if !ok {
		return
	}

	cacheData, isFoundAndNotExpired := t.cache.Get(cacheRequestData.Identifier)
	response.Data = thread.CacheResponseData{
		Identifier: cacheRequestData.Identifier,
		Data:       cacheData,
	}
	response.Success = isFoundAndNotExpired && (cacheData != nil)
}

func (t *Thread) processPingCache(request *thread.Request, response *thread.Response) {

	if t.config.Debug {
		t.logger.Println("received ping over c9")
	}
	response.Success = true
}
