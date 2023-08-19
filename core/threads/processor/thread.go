package processor

import (
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango/module"
	"github.com/GabeCordo/mango/processor"
	"github.com/GabeCordo/mango/threads"
)

func (thread *Thread) Setup() {
	thread.accepting = true
}

func (thread *Thread) Start() {

	// INCOMING REQUESTS
	go func() {
		// request coming from http_client
		for request := range thread.C5 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = threads.HttpClient
			thread.processRequest(&request)
		}
	}()

	go func() {
		// request coming from http_processor
		for request := range thread.C7 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
			request.Source = threads.HttpProcessor
			thread.processRequest(&request)
		}
	}()

	// RESPONSE THREADS

	go func() {
		// response coming from the supervisor thread
		for response := range thread.C14 {
			thread.SupervisorResponseTable.Write(response.Nonce, response)
		}
	}()

	go func() {
		// response coming from the database thread
		for response := range thread.C12 {
			thread.DatabaseResponseTable.Write(response.Nonce, response)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) request(dest threads.Module, request any) error {
	switch dest {
	case threads.Supervisor:
		thread.C13 <- (request).(common.SupervisorRequest)
	case threads.Database:
		thread.C11 <- (request).(threads.DatabaseRequest)
	default:
		return threads.BadRequestType
	}

	return nil
}

func (thread *Thread) respond(source threads.Module, response *common.ProcessorResponse) error {
	switch source {
	case threads.HttpClient:
		thread.C6 <- *response
	case threads.HttpProcessor:
		thread.C8 <- *response
	default:
		return threads.BadResponseType
	}

	return nil
}

func (thread *Thread) processRequest(request *common.ProcessorRequest) {

	response := &common.ProcessorResponse{Nonce: request.Nonce, Error: nil}

	switch request.Action {
	case common.ProcessorGet:
		response.Data = thread.processorGet()
	case common.ProcessorAdd:
		cfg := (request.Data).(processor.Config)
		response.Error = thread.processorAdd(&cfg)
	case common.ProcessorRemove:
		cfg := (request.Data).(module.Config)
		thread.processorRemove(&cfg)
	case common.ProcessorModuleGet:
		response.Data = thread.getModules()
	case common.ProcessorModuleAdd:
		cfg := (request.Data).(module.Config)
		response.Error = thread.addModule(request.Identifiers.Processor, &cfg)
	case common.ProcessorModuleDelete:
		response.Error = thread.deleteModule(request.Identifiers.Processor, request.Identifiers.Module)
	case common.ProcessorModuleMount:
		response.Error = thread.mountModule(request.Identifiers.Module)
	case common.ProcessorModuleUnmount:
		response.Error = thread.unmountModule(request.Identifiers.Module)
	case common.ProcessorClusterGet:
		response.Data, response.Error = thread.getClusters(request.Identifiers.Module)
	case common.ProcessorClusterMount:
		response.Error = thread.mountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
	case common.ProcessorClusterUnmount:
		response.Error = thread.unmountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
	case common.ProcessorSupervisorFetch:
		response.Data, response.Error = thread.fetchSupervisor(request)
	case common.ProcessorSupervisorCreate:
		response.Data, response.Error = thread.createSupervisor(request)
	case common.ProcessorSupervisorUpdate:
		response.Error = thread.updateSupervisor(request)
	default:
		response.Error = threads.UnknownRequest
	}

	response.Success = response.Error == nil
	thread.respond(request.Source, response)
	thread.wg.Done()
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait()
}
