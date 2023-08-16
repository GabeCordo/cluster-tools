package supervisor

import "github.com/GabeCordo/etl/core/threads/common"

func (thread *Thread) Setup() {
	thread.accepting = true
}

func (thread *Thread) Start() {

	// INCOMING REQUESTS

	go func() {
		// request coming from http_server
		for request := range thread.C14 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
			thread.ProcessRequest(&request)
		}
	}()

	// INCOMING RESPONSES

	go func() {
		// request coming from http_server
		for response := range thread.C8 {
			// if this doesn't spawn its own thread we will be left waiting
			thread.DatabaseResponseTable.Write(response.Nonce, response)
		}
	}()

	go func() {
		// request coming from http_server
		for response := range thread.C10 {
			// if this doesn't spawn its own thread we will be left waiting
			thread.CacheResponseTable.Write(response.Nonce, response)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) ProcessRequest(request *common.SupervisorRequest) {

	switch request.Action {
	case common.SupervisorCreate:
		thread.createSupervisor(
			request.Identifiers.Module, request.Identifiers.Module, request.Identifiers.Config)
	case common.SupervisorError:
		GetRegistryInstance().Get(request.Identifiers.Supervisor)
	case common.SupervisorClose:
	}
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait() // don't tear down until all the requests have been processed
}
