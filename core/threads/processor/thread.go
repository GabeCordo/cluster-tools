package processor

import (
	"github.com/GabeCordo/cluster-tools/core/interfaces/module"
	"github.com/GabeCordo/cluster-tools/core/interfaces/processor"
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
		// request coming from http_client
		for request := range thread.C5 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.HttpClient
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
			request.Source = common.HttpProcessor
			thread.processRequest(&request)
		}
	}()

	go func() {
		// request coming from http_processor
		for request := range thread.C18 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
			request.Source = common.Scheduler
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

func (thread *Thread) request(dest common.Module, request any) error {
	switch dest {
	case common.Supervisor:
		thread.C13 <- (request).(common.SupervisorRequest)
	case common.Database:
		thread.C11 <- (request).(common.DatabaseRequest)
	default:
		return common.BadRequestType
	}

	return nil
}

func (thread *Thread) respond(source common.Module, response *common.ProcessorResponse) error {
	switch source {
	case common.HttpClient:
		thread.C6 <- *response
	case common.HttpProcessor:
		thread.C8 <- *response
	case common.Scheduler:
		thread.C19 <- *response
	default:
		return common.BadResponseType
	}

	return nil
}

func (thread *Thread) processRequest(request *common.ProcessorRequest) {

	response := &common.ProcessorResponse{Nonce: request.Nonce, Error: nil}

	switch request.Action {
	case common.ProcessorPing:
		response.Error = thread.ping()
	case common.ProcessorGet:
		response.Data = thread.processorGet()
	case common.ProcessorAdd:
		cfg := (request.Data).(processor.Config)
		response.Error = thread.processorAdd(&cfg)
	case common.ProcessorRemove:
		cfg := (request.Data).(processor.Config)
		response.Error = thread.processorRemove(&cfg)
	case common.ProcessorModuleGet:
		response.Data = thread.getModules()
	case common.ProcessorModuleAdd:
		cfg := (request.Data).(module.Config)
		response.Error = thread.addModule(request.Identifiers.Processor, &cfg)
	case common.ProcessorModuleRemove:
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
	case common.ProcessorSupervisorGet:
		response.Data, response.Error = thread.getSupervisor(request)
	case common.ProcessorSupervisorCreate:
		response.Data, response.Error = thread.createSupervisor(request)
	case common.ProcessorSupervisorUpdate:
		response.Error = thread.updateSupervisor(request)
	case common.ProcessorSupervisorLog:
		response.Error = thread.logSupervisor(request)
	default:
		response.Error = common.UnknownRequest
	}

	response.Success = response.Error == nil
	thread.respond(request.Source, response)
	thread.wg.Done()
}

func (thread *Thread) ping() error {

	thread.Logger.Println("received ping over C5")

	// TEST DB CHANNELS

	dbRequest := common.DatabaseRequest{
		Action: common.DatabaseLowerPing,
		Nonce:  rand.Uint32(),
	}
	thread.C11 <- dbRequest

	rsp, didTimeout := multithreaded.SendAndWait(thread.DatabaseResponseTable, dbRequest.Nonce, thread.config.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	dbResponse := (rsp).(common.DatabaseResponse)
	if dbResponse.Error != nil {
		return dbResponse.Error
	}

	thread.Logger.Println("received ping over C12")

	// TEST SUPERVISOR CHANNELS

	supRequest := common.SupervisorRequest{
		Action: common.SupervisorPing,
		Nonce:  rand.Uint32(),
	}
	thread.C13 <- supRequest

	rsp, didTimeout = multithreaded.SendAndWait(thread.SupervisorResponseTable, supRequest.Nonce, thread.config.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	supResponse := (rsp).(common.SupervisorResponse)
	return supResponse.Error
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait()
}
