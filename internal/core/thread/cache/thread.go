package cache

import (
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
	"time"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	thread.SetupListener(t.C9, t.C10, &t.accepting, &t.wg, thread.Cache, t.Handle)

	thread.SetupListener(t.C24, t.C25, &t.accepting, &t.wg, thread.Cache, t.Handle)

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

func (t *Thread) Respond(response *thread.Response) {

	t.C10 <- *response
}

func (t *Thread) Handle(request *thread.Request, response *thread.Response) {

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
