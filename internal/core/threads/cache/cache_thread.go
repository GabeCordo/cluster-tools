package cache

import (
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
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
		// request from http_server
		for request := range thread.C24 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.HttpClient
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

func (thread *Thread) Respond(response *common.ThreadResponse) {

	thread.C10 <- *response
}

func (thread *Thread) ProcessIncomingRequest(request *common.ThreadRequest) {
	if request.Action == common.CreateAction {
		thread.processSaveRequest(request)
	} else if request.Action == common.GetAction {
		thread.processLoadRequest(request)
	} else if request.Action == common.PingAction {
		thread.processPingCache(request)
	}

	thread.wg.Done()
}

func (thread *Thread) Teardown() {
	thread.accepting = false
}
