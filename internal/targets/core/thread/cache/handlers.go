package cache

import (
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/thread"
)

// processSaveRequest
// will insert or override an existing cache record
func (t *Thread) processSaveRequest(request *thread.Request, response *thread.Response) {

	cacheRequestData, ok := (request.Data).(thread.CacheRequestData)
	if !ok {
		response.Success = false
		response.Error = thread.BadRequestType
		return
	}

	id, err := t.useCase.Save(cacheRequestData.Identifier, cacheRequestData.Data, cacheRequestData.ExpiresIn)
	if err != nil {
		response.Success = false
		response.Error = err
		return
	}

	response.Success = true
	response.Data = thread.CacheResponseData{
		Identifier: cacheRequestData.Identifier,
		Data:       id,
	}
}

func (t *Thread) processLoadRequest(request *thread.Request, response *thread.Response) {

	cacheRequestData, ok := (request.Data).(thread.CacheRequestData)
	if !ok {
		response.Success = false
		response.Error = thread.BadRequestType
		return
	}

	data, err := t.useCase.Load(cacheRequestData.Identifier)
	if err != nil {
		response.Success = false
		response.Error = err
		return
	}

	response.Success = true
	response.Data = thread.CacheResponseData{
		Identifier: cacheRequestData.Identifier,
		Data:       data,
	}
}

func (t *Thread) processPingCache(request *thread.Request, response *thread.Response) {

	if t.config.Debug {
		t.logger.Println("received ping over c9")
	}
	response.Success = true
}
