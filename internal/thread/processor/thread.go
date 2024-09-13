package processor

import (
	"github.com/GabeCordo/cluster-tools/internal/processor"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"time"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	// INCOMING REQUESTS
	go func() {
		// request coming from client
		for request := range t.C5 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			request.Source = thread.HttpClient
			t.processRequest(&request)
		}
	}()

	go func() {
		// request coming from processor
		for request := range t.C7 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			// if this doesn't spawn its own t we will be left waiting
			request.Source = thread.HttpProcessor
			t.processRequest(&request)
		}
	}()

	go func() {
		// request coming from processor
		for request := range t.C18 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			// if this doesn't spawn its own t we will be left waiting
			request.Source = thread.Scheduler
			t.processRequest(&request)
		}
	}()

	// RESPONSE THREADS

	go func() {
		// response coming from the supervisor t
		for response := range t.C14 {
			t.SupervisorResponseTable.Write(response.Nonce, response)
		}
	}()

	go func() {
		// response coming from the database t
		for response := range t.C12 {
			t.DatabaseResponseTable.Write(response.Nonce, response)
		}
	}()

	// PROCESSOR PROBE LOOP

	go func() {
		sleepDuration := time.Duration(t.config.ProbeEvery) * time.Second

		for {
			t.processorPing()
			time.Sleep(sleepDuration)
		}
	}()
}

func (t *Thread) request(dest thread.Module, request *thread.Request) error {
	switch dest {
	case thread.Supervisor:
		t.C13 <- *request
	case thread.Database:
		t.C11 <- *request
	default:
		return thread.BadRequestType
	}

	return nil
}

func (t *Thread) respond(source thread.Module, response *thread.Response) error {
	switch source {
	case thread.HttpClient:
		t.C6 <- *response
	case thread.HttpProcessor:
		t.C8 <- *response
	case thread.Scheduler:
		t.C19 <- *response
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
		case thread.ProcessorRecord:
			response.Data = t.processorGet()
		case thread.ModuleRecord:
			response.Data = t.getModules()
		case thread.ClusterRecord:
			response.Data, response.Error = t.getClusters(request.Identifiers.Module)
		case thread.SupervisorRecord:
			response.Data, response.Error = t.getSupervisor(request)
		default:
			response.Error = thread.UnknownRequest
		}
	case thread.CreateAction:
		switch request.Type {
		case thread.ProcessorRecord:
			cfg := (request.Data).(processor.Config)
			response.Error = t.processorAdd(&cfg)
		case thread.ModuleRecord:
			cfg := (request.Data).(processor.ModuleConfig)
			response.Error = t.addModule(request.Identifiers.Processor, &cfg)
		case thread.SupervisorRecord:
			response.Data, response.Error = t.createSupervisor(request)
		default:
			response.Error = thread.UnknownRequest
		}
	case thread.DeleteAction:
		switch request.Type {
		case thread.ProcessorRecord:
			cfg := (request.Data).(processor.Config)
			response.Error = t.processorRemove(&cfg)
		case thread.ModuleRecord:
			response.Error = t.deleteModule(request.Identifiers.Processor, request.Identifiers.Module)
		default:
			response.Error = thread.UnknownRequest
		}
	case thread.UpdateAction:
		switch request.Type {
		case thread.SupervisorRecord:
			response.Error = t.updateSupervisor(request)
		default:
			response.Error = thread.UnknownRequest
		}
	case thread.MountAction:
		switch request.Type {
		case thread.ModuleRecord:
			response.Error = t.mountModule(request.Identifiers.Module)
		case thread.ClusterRecord:
			response.Error = t.mountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
		default:
			response.Error = thread.UnknownRequest
		}
	case thread.UnMountAction:
		switch request.Type {
		case thread.ModuleRecord:
			response.Error = t.unmountModule(request.Identifiers.Module)
		case thread.ClusterRecord:
			response.Error = t.unmountCluster(request.Identifiers.Module, request.Identifiers.Cluster)
		default:
			response.Error = thread.UnknownRequest
		}
	case thread.LogAction:
		switch request.Type {
		case thread.SupervisorRecord:
			response.Error = t.logSupervisor(request)
		default:
			response.Error = thread.UnknownRequest
		}
	default:
		response.Error = thread.UnknownRequest
	}

	response.Success = response.Error == nil
	t.respond(request.Source, response)
	t.wg.Done()
}

func (t *Thread) Teardown() {
	t.accepting = false
	t.wg.Wait()
}
