package processor

import (
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"time"
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

	// PROCESSOR PROBE LOOP

	go func() {
		sleepDuration := time.Duration(thread.config.ProbeEvery) * time.Second

		for {
			thread.processorPing()
			time.Sleep(sleepDuration)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) request(dest common.Module, request *common.ThreadRequest) error {
	switch dest {
	case common.Supervisor:
		thread.C13 <- *request
	case common.Database:
		thread.C11 <- *request
	default:
		return common.BadRequestType
	}

	return nil
}

func (thread *Thread) respond(source common.Module, response *common.ThreadResponse) error {
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

func (thread *Thread) processRequest(request *common.ThreadRequest) {

	response := &common.ThreadResponse{Nonce: request.Nonce, Error: nil}

	switch request.Action {
	case common.GetAction:
		switch request.Type {
		case common.ProcessorRecord:
			response.Data = thread.processorGet()
		case common.ModuleRecord:
			response.Data = thread.getModules()
		case common.ClusterRecord:
			response.Data, response.Error = thread.getClusters(request.Identifiers.Module)
		case common.SupervisorRecord:
			response.Data, response.Error = thread.getSupervisor(request)
		default:
			response.Error = common.UnknownRequest
		}
	case common.CreateAction:
		switch request.Type {
		case common.ProcessorRecord:
			cfg := (request.Data).(interfaces.ProcessorConfig)
			response.Error = thread.processorAdd(&cfg)
		case common.ModuleRecord:
			cfg := (request.Data).(interfaces.ModuleConfig)
			response.Error = thread.addModule(request.Identifiers.Processor, &cfg)
		case common.SupervisorRecord:
			response.Data, response.Error = thread.createSupervisor(request)
		default:
			response.Error = common.UnknownRequest
		}
	case common.DeleteAction:
		switch request.Type {
		case common.ProcessorRecord:
			cfg := (request.Data).(interfaces.ProcessorConfig)
			response.Error = thread.processorRemove(&cfg)
		case common.ModuleRecord:
			response.Error = thread.deleteModule(request.Identifiers.Processor, request.Identifiers.Module)
		default:
			response.Error = common.UnknownRequest
		}
	case common.UpdateAction:
		switch request.Type {
		case common.SupervisorRecord:
			response.Error = thread.updateSupervisor(request)
		default:
			response.Error = common.UnknownRequest
		}
	case common.MountAction:
		switch request.Type {
		case common.ModuleRecord:
			response.Error = thread.mountModule(request.Identifiers.Module)
		case common.ClusterRecord:
			response.Error = thread.mountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
		default:
			response.Error = common.UnknownRequest
		}
	case common.UnMountAction:
		switch request.Type {
		case common.ModuleRecord:
			response.Error = thread.unmountModule(request.Identifiers.Module)
		case common.ClusterRecord:
			response.Error = thread.unmountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
		default:
			response.Error = common.UnknownRequest
		}
	case common.LogAction:
		switch request.Type {
		case common.SupervisorRecord:
			response.Error = thread.logSupervisor(request)
		default:
			response.Error = common.UnknownRequest
		}
	default:
		response.Error = common.UnknownRequest
	}

	response.Success = response.Error == nil
	thread.respond(request.Source, response)
	thread.wg.Done()
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait()
}
