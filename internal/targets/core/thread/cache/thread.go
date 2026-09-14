package cache

import (
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/thread"
)

func (t *Thread) Setup() {

}

func (t *Thread) Start() {

	running := true

	go func(running *bool) {
		// cleaning the t of expired records
		for *running {
			time.Sleep(1 * time.Minute)
			// every minute, attempt to clean the t by removing any records that
			// may have expired since we last checked
			t.useCase.CacheComponent.Clean()
		}
	}(&running)

	var iReq *thread.Request
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-t.channels.c9:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					t.channels.c10 <- oRsp
				}
			}
		case iReq = <-t.channels.c24:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					t.channels.c25 <- oRsp
				}
			}
		case <-t.channels.close:
			{
				// shutting down the cache thread
				running = false
				break
			}
		}
		oRsp = nil
	}
}

func (t *Thread) handleRequest(request *thread.Request) (response *thread.Response) {

	response = thread.NewResponse(thread.Messenger)

	switch request.Action {
	case thread.CreateAction:
		{
			t.processSaveRequest(request, response)
		}
	case thread.GetAction:
		{
			t.processLoadRequest(request, response)
		}
	case thread.PingAction:
		{
			t.processPingCache(request, response)
		}
	default:
		{
			response.Error = thread.UnknownRequest
		}
	}

	return response
}

func (t *Thread) Teardown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
