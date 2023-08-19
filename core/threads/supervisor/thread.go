package supervisor

import (
	"github.com/GabeCordo/mango-core/core/components/supervisor"
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango/threads"
)

func (thread *Thread) Setup() {
	thread.accepting = true
}

func (thread *Thread) Start() {

	// INCOMING REQUESTS

	go func() {
		for request := range thread.C13 {
			if !thread.accepting {
				break
			}

			request.Source = threads.Processor
			thread.processRequest(&request)
		}
	}()

	// INCOMING RESPONSES

	go func() {
		// response coming from database thread
		for response := range thread.C16 {
			// if this doesn't spawn its own thread we will be left waiting
			thread.DatabaseResponseTable.Write(response.Nonce, response)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) respond(dst threads.Module, response *common.SupervisorResponse) error {
	switch dst {
	case threads.Processor:
		thread.C14 <- *response
	default:
		return threads.BadResponseType
	}

	return nil
}

func (thread *Thread) processRequest(request *common.SupervisorRequest) {

	response := &common.SupervisorResponse{Nonce: request.Nonce, Error: nil}

	switch request.Action {
	case common.SupervisorCreate:
		response.Data, response.Error = thread.createSupervisor(
			request.Identifiers.Processor, request.Identifiers.Module, request.Identifiers.Module, request.Identifiers.Config)
	case common.SupervisorUpdate:
		s := (request.Data).(supervisor.Supervisor)
		response.Error = thread.updateSupervisor(&s)
	default:
		response.Error = threads.BadRequestType
	}

	response.Success = response.Error == nil
	thread.respond(request.Source, response)
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait() // don't tear down until all the requests have been processed
}
