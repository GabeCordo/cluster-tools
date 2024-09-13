package cache

import (
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"time"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	go func() {
		// request from http_server
		for request := range t.C9 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			request.Source = thread.HttpProcessor
			t.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// request from http_server
		for request := range t.C24 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			request.Source = thread.HttpClient
			t.ProcessIncomingRequest(&request)
		}
	}()

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

func (t *Thread) ProcessIncomingRequest(request *thread.Request) {
	if request.Action == thread.CreateAction {
		t.processSaveRequest(request)
	} else if request.Action == thread.GetAction {
		t.processLoadRequest(request)
	} else if request.Action == thread.PingAction {
		t.processPingCache(request)
	}

	t.wg.Done()
}

func (t *Thread) Teardown() {
	t.accepting = false
	t.wg.Wait()
}
