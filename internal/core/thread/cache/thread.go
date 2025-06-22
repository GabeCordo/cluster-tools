package cache

import (
	"time"

	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	thread.SetupListener(t.channels.c9, t.channels.c10, &t.accepting, &t.wg, thread.Cache, t.HandleRequest)

	thread.SetupListener(t.channels.c24, t.channels.c25, &t.accepting, &t.wg, thread.Cache, t.HandleRequest)

	// RUNTIME

	go func() {
		// cleaning the t of expired records
		for t.accepting {
			time.Sleep(1 * time.Minute)
			// every minute, attempt to clean the t by removing any records that
			// may have expired since we last checked
			t.cache.Clean()
		}
	}()
}

func (t *Thread) HandleRequest(request *thread.Request, response *thread.Response) {

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
			t.logger.Warn(thread.UnknownRequest.Error())
		}
	}
}

func (t *Thread) Teardown() {
	t.accepting = false
	t.wg.Wait()
}
