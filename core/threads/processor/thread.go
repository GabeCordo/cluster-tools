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
		for request := range thread.C12 {
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
		// request coming from http_server
		for response := range thread.C15 {
			// if this doesn't spawn its own thread we will be left waiting
			thread.SupervisorResponseTable.Write(response.Nonce, response)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) respond(source threads.Module, response *common.ProcessorResponse) {
	switch source {
	case threads.HttpClient:
		thread.C6 <- *response
	case threads.HttpProcessor:
		thread.C13 <- *response
	}
}

func (thread *Thread) processRequest(request *common.ProcessorRequest) {
	switch request.Action {
	case common.ProcessorGet:
		processors := thread.processorGet()
		thread.respond(request.Source, &common.ProcessorResponse{Success: true, Data: processors, Nonce: request.Nonce})
	case common.ProcessorAdd:
		thread.processorAdd(request.Data.Module)
	case common.ProcessorRemove:
		thread.processorRemove(request.Data.Module)
	case common.ProcessorModuleGet:
		modules := thread.getModules()
		thread.respond(request.Source, &common.ProcessorResponse{Success: true, Data: modules, Nonce: request.Nonce})
	case common.ProcessorModuleMount:
		thread.mountModule(request.Identifiers.Module)
	case common.ProcessorModuleUnmount:
		thread.unmountModule(request.Identifiers.Module)
	case common.ProcessorClusterGet:
		clusters, found := thread.getClusters(request.Identifiers.Module)
		thread.respond(request.Source, &common.ProcessorResponse{Success: found, Data: clusters, Nonce: request.Nonce})
	case common.ProcessorClusterMount:
		thread.mountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
	case common.ProcessorClusterUnmount:
		thread.unmountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
	}

	thread.wg.Done()
}

func (thread *Thread) Teardown() {
	thread.accepting = false
	thread.wg.Wait()
}
