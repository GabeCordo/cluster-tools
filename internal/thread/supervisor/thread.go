package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/message/log"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"strconv"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	// INCOMING REQUESTS

	go func() {
		for request := range t.C13 {
			if !t.accepting {
				break
			}

			request.Source = thread.Processor
			t.processRequest(&request)
		}
	}()

	// INCOMING RESPONSES

	go func() {
		// response coming from database thread
		for response := range t.C16 {
			// if this doesn't spawn its own thread we will be left waiting
			t.DatabaseResponseTable.Write(response.Nonce, response)
		}
	}()
}

func (t *Thread) respond(dst thread.Module, response *thread.Response) error {
	switch dst {
	case thread.Processor:
		t.C14 <- *response
	default:
		return thread.BadResponseType
	}

	return nil
}

func (t *Thread) processRequest(request *thread.Request) {

	response := &thread.Response{Nonce: request.Nonce, Error: nil}

	switch request.Action {
	case thread.GetAction:
		switch request.Type {
		case thread.SupervisorRecord:
			f := &database.Filter{
				Module:     request.Identifiers.Module,
				Cluster:    request.Identifiers.Cluster,
				Identifier: strconv.FormatUint(request.Identifiers.Supervisor, 10),
			}
			response.Data, response.Error = t.getSupervisor(f)
		default:
			response.Error = thread.BadRequestType
		}
	case thread.CreateAction:
		switch request.Type {
		case thread.SupervisorRecord:
			metadata, success := (request.Data).(map[string]string)
			if !success {
				response.Error = errors.New("SupervisorCreate expected a map[string]string data type")
			} else {
				response.Data, response.Error = t.createSupervisor(
					request.Identifiers.Processor, request.Identifiers.Module,
					request.Identifiers.Config, request.Identifiers.Config,
					metadata)
			}
		default:
			response.Error = thread.BadRequestType
		}
	case thread.UpdateAction:
		switch request.Type {
		case thread.SupervisorRecord:
			s := (request.Data).(*supervisor.Supervisor)
			response.Error = t.updateSupervisor(s)
		default:
			response.Error = thread.BadRequestType
		}
	case thread.LogAction:
		switch request.Type {
		case thread.SupervisorRecord:
			l := (request.Data).(*log.Log)
			response.Error = t.logSupervisor(l)
		default:
			response.Error = thread.BadRequestType
		}
	default:
		response.Error = thread.BadRequestType
	}

	response.Success = response.Error == nil
	t.respond(request.Source, response)
}

func (t *Thread) Teardown() {
	t.accepting = false
	t.wg.Wait() // don't tear down until all the requests have been processed
}
