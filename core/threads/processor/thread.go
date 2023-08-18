package processor

import (
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl/core/threads/common"
)

func (thread *Thread) Setup() {
	thread.accepting = true
}

func (thread *Thread) Start() {

	// INCOMING REQUESTS
	go func() {
		// request coming from http_server
		for request := range thread.C5 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
			thread.processRequest(&request)
		}
	}()

	go func() {
		// request coming from http_server
		for request := range thread.C7 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			// if this doesn't spawn its own thread we will be left waiting
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

func (thread *Thread) respond(source threads.Module, response *common.ProcessorResponse) {
	switch source {
	case threads.HttpClient:
		thread.C6 <- *response
	case threads.HttpProcessor:
		thread.C8 <- *response
	}
}

func (thread *Thread) processRequest(request *common.ProcessorRequest) {
	switch request.Action {
	case common.ProcessorGet:
		processors := thread.processorGet()
		thread.respond(request.Source, &common.ProcessorResponse{Success: true, Data: processors, Nonce: request.Nonce})
	case common.ProcessorAdd:
		err := thread.processorAdd(&request.Data.Processor)
		thread.respond(request.Source, &common.ProcessorResponse{Success: err == nil, Error: err, Nonce: request.Nonce})
	case common.ProcessorRemove:
		thread.processorRemove(request.Data.Module)
	case common.ProcessorModuleGet:
		modules := thread.getModules()
		thread.respond(request.Source, &common.ProcessorResponse{Success: true, Data: modules, Nonce: request.Nonce})
	case common.ProcessorModuleAdd:
		err := thread.addModule(request.Identifiers.Processor, &request.Data.Module)
		thread.respond(request.Source, &common.ProcessorResponse{Success: err == nil, Error: err, Nonce: request.Nonce})
	case common.ProcessorModuleDelete:
		err := thread.deleteModule(request.Identifiers.Processor, request.Identifiers.Module)
		thread.respond(request.Source, &common.ProcessorResponse{Success: err == nil, Error: err, Nonce: request.Nonce})
	case common.ProcessorModuleMount:
		err := thread.mountModule(request.Identifiers.Module)
		thread.respond(request.Source, &common.ProcessorResponse{Success: err == nil, Error: err, Nonce: request.Nonce})
	case common.ProcessorModuleUnmount:
		err := thread.unmountModule(request.Identifiers.Module)
		thread.respond(request.Source, &common.ProcessorResponse{Success: err == nil, Error: err, Nonce: request.Nonce})
	case common.ProcessorClusterGet:
		clusters, found := thread.getClusters(request.Identifiers.Module)
		thread.respond(request.Source, &common.ProcessorResponse{Success: found, Data: clusters, Nonce: request.Nonce})
	case common.ProcessorClusterMount:
		err := thread.mountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
		thread.respond(request.Source, &common.ProcessorResponse{Success: err == nil, Error: err, Nonce: request.Nonce})
	case common.ProcessorClusterUnmount:
		err := thread.unmountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
		thread.respond(request.Source, &common.ProcessorResponse{Success: err == nil, Error: err, Nonce: request.Nonce})
	}

	thread.wg.Done()
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait()
}
