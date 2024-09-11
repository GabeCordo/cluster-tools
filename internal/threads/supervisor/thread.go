package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
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

			request.Source = common.Processor
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

func (thread *Thread) respond(dst common.Module, response *common.ThreadResponse) error {
	switch dst {
	case common.Processor:
		thread.C14 <- *response
	default:
		return common.BadResponseType
	}

	return nil
}

func (thread *Thread) processRequest(request *common.ThreadRequest) {

	response := &common.ThreadResponse{Nonce: request.Nonce, Error: nil}

	switch request.Action {
	case common.GetAction:
		switch request.Type {
		case common.SupervisorRecord:
			f := &supervisor.Filter{
				Module:  request.Identifiers.Module,
				Cluster: request.Identifiers.Cluster,
				Id:      request.Identifiers.Supervisor,
			}
			response.Data, response.Error = thread.getSupervisor(f)
		default:
			response.Error = common.BadRequestType
		}
	case common.CreateAction:
		switch request.Type {
		case common.SupervisorRecord:
			metadata, success := (request.Data).(map[string]string)
			if !success {
				response.Error = errors.New("SupervisorCreate expected a map[string]string data type")
			} else {
				response.Data, response.Error = thread.createSupervisor(
					request.Identifiers.Processor, request.Identifiers.Module,
					request.Identifiers.Config, request.Identifiers.Config,
					metadata)
			}
		default:
			response.Error = common.BadRequestType
		}
	case common.UpdateAction:
		switch request.Type {
		case common.SupervisorRecord:
			s := (request.Data).(*supervisor.Supervisor)
			response.Error = thread.updateSupervisor(s)
		default:
			response.Error = common.BadRequestType
		}
	case common.LogAction:
		switch request.Type {
		case common.SupervisorRecord:
			log := (request.Data).(*supervisor.Log)
			response.Error = thread.logSupervisor(log)
		default:
			response.Error = common.BadRequestType
		}
	default:
		response.Error = common.BadRequestType
	}

	response.Success = response.Error == nil
	thread.respond(request.Source, response)
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait() // don't tear down until all the requests have been processed
}
