package processor

import (
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	// INCOMING REQUESTS
	thread.SetupListener(t.C5, t.C6, &t.accepting, &t.wg, thread.Processor, t.Handle)

	thread.SetupListener(t.C7, t.C8, &t.accepting, &t.wg, thread.Processor, t.Handle)

	thread.SetupListener(t.C18, t.C19, &t.accepting, &t.wg, thread.Processor, t.Handle)

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

	// TODO : stopping probe feature for testing
	//go func() {
	//	sleepDuration := time.Duration(t.pipeline.ProbeEvery) * time.Second
	//
	//	for {
	//		t.processorPing()
	//		time.Sleep(sleepDuration)
	//	}
	//}()
}

func (t *Thread) Handle(request *thread.Request, response *thread.Response) {

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				response.Data = t.processorGet()
			case thread.ModuleRecord:
				response.Data = t.getModules()
			case thread.FunctionRecord:
				response.Data, response.Error = t.getClusters(request.Identifiers.Module)
			case thread.SupervisorRecord:
				response.Data, response.Error = t.getSupervisor(request)
			default:
				response.Error = thread.UnknownRequest
			}
		}
	case thread.CreateAction:
		{
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
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				cfg := (request.Data).(processor.Config)
				response.Error = t.processorRemove(&cfg)
			case thread.ModuleRecord:
				response.Error = t.deleteModule(request.Identifiers.Processor, request.Identifiers.Module)
			default:
				response.Error = thread.UnknownRequest
			}
		}
	case thread.UpdateAction:
		{
			switch request.Type {
			case thread.SupervisorRecord:
				response.Error = t.updateSupervisor(request)
			default:
				response.Error = thread.UnknownRequest
			}
		}
	case thread.MountAction:
		{
			switch request.Type {
			case thread.ModuleRecord:
				response.Error = t.mountModule(request.Identifiers.Module)
			case thread.FunctionRecord:
				response.Error = t.mountCluster(request.Identifiers.Module, request.Identifiers.Function)
			default:
				response.Error = thread.UnknownRequest
			}
		}
	case thread.UnMountAction:
		{
			switch request.Type {
			case thread.ModuleRecord:
				response.Error = t.unmountModule(request.Identifiers.Module)
			case thread.FunctionRecord:
				response.Error = t.unmountCluster(request.Identifiers.Module, request.Identifiers.Function)
			default:
				response.Error = thread.UnknownRequest
			}
		}
	case thread.LogAction:
		{
			switch request.Type {
			case thread.SupervisorRecord:
				response.Error = t.logSupervisor(request)
			default:
				response.Error = thread.UnknownRequest
			}
		}
	default:
		{
			response.Error = thread.UnknownRequest
		}
	}

	response.Success = response.Error == nil
}

func (t *Thread) Teardown() {
	t.accepting = false
	t.wg.Wait()
}
