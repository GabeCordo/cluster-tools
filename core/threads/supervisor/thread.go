package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
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

func (thread *Thread) respond(dst common.Module, response *common.SupervisorResponse) error {
	switch dst {
	case common.Processor:
		thread.C14 <- *response
	default:
		return common.BadResponseType
	}

	return nil
}

func (thread *Thread) processRequest(request *common.SupervisorRequest) {

	response := &common.SupervisorResponse{Nonce: request.Nonce, Error: nil}

	switch request.Action {
	case common.SupervisorPing:
		response.Error = thread.ping()
	case common.SupervisorGet:
		response.Data, response.Error = thread.getSupervisor(request.Identifiers.Supervisor)
	case common.SupervisorCreate:
		metadata, success := (request.Data).(map[string]string)
		if !success {
			response.Error = errors.New("SupervisorCreate expected a map[string]string data type")
		} else {
			response.Data, response.Error = thread.createSupervisor(
				request.Identifiers.Processor, request.Identifiers.Module,
				request.Identifiers.Config, request.Identifiers.Config,
				metadata)
		}
	case common.SupervisorUpdate:
		s := (request.Data).(*supervisor.Supervisor)
		response.Error = thread.updateSupervisor(s)
	case common.SupervisorLog:
		log := (request.Data).(*supervisor.Log)
		response.Error = thread.logSupervisor(log)
	default:
		response.Error = common.BadRequestType
	}

	response.Success = response.Error == nil
	thread.respond(request.Source, response)
}

func (thread *Thread) ping() error {

	thread.Logger.Println("received ping over C13")

	// TEST DATABASE CHANNELS

	request := common.DatabaseRequest{
		Action: common.DatabaseLowerPing,
		Nonce:  rand.Uint32(),
	}
	thread.C15 <- request

	rsp, didTimeout := multithreaded.SendAndWait(thread.DatabaseResponseTable, request.Nonce, thread.config.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(common.DatabaseResponse)
	return response.Error
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait() // don't tear down until all the requests have been processed
}
