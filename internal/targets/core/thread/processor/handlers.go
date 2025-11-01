package processor

import (
	processor2 "github.com/FortifiedCode/flock/internal/targets/core/component/processor"
	thread2 "github.com/FortifiedCode/flock/internal/targets/core/thread"
	"github.com/FortifiedCode/plover"
)

//////////////////////////////////////////////////////////////////////////////////////////
////							~~~ HANDLE REQUESTS ~~~
//////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////////////////
////							Getter Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleGetProcessor(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Data = t.useCases.GetProcessors()
	(*response).Success = true
}

func (t *Thread) handleGetModule(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Data = t.useCases.GetModules()
	(*response).Success = true
}

func (t *Thread) handleGetFunctions(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Data, (*response).Error = t.useCases.GetFunctions(request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleGetRun(request *thread2.Request, response **thread2.Response) {

	t.requestStore[request.Nonce] = request

	// processor -> all runner ids on the processor
	//	-	id
	// /module -> all runner ids of the module, on the processor
	//	-	id
	//	-	status?
	// processor/module/cluster -> all ids of that cluster, of the module, on the processor
	//	-	id
	//	-	status?
	//	-	num processed?
	// id -> the entire record of the runner
	//	-	full information
	thread2.AsyncGetRun(t.channels.c13, request.Nonce,
		request.Identifiers.Namespace, request.Identifiers.Pipeline, request.Identifiers.Supervisor)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Create Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleCreateProcessor(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)

	cfg, ok := (request.Data).(processor2.Config)
	if !ok {
		(*response).Success = false
		(*response).Error = thread2.BadRequestType
		return
	}

	(*response).Error = t.useCases.AddProcessor(&cfg)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleCreateModule(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)

	cfg, ok := (request.Data).(*plover.ModuleIR)
	if !ok {
		(*response).Success = false
		(*response).Error = thread2.BadRequestType
		return
	}

	(*response).Error = t.useCases.AddModule(request.Identifiers.Processor, cfg)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleCreateRun(request *thread2.Request, response **thread2.Response) {

	// fetch the pipeline from the database is async
	t.requestStore[request.Nonce] = request
	thread2.AsyncGetPipeline(t.channels.c11, request.Nonce,
		request.Identifiers.Namespace, request.Identifiers.Pipeline)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Delete Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleDeleteProcessor(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)

	cfg, ok := (request.Data).(processor2.Config)
	if !ok {
		(*response).Success = false
		(*response).Error = thread2.BadRequestType
		return
	}

	(*response).Error = t.useCases.DeleteProcessor(&cfg)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleDeleteModule(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Error = t.useCases.DeleteModule(request.Identifiers.Processor, request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleDeleteRun(request *thread2.Request, response **thread2.Response) {

	t.requestStore[request.Nonce] = request
	thread2.AsyncStopRunToRunner(t.channels.c13, request)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Update Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleUpdateRun(request *thread2.Request, response **thread2.Response) {

	t.requestStore[request.Nonce] = request
	thread2.AsyncUpdateRunToRunner(t.channels.c13, request)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Mount Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleMountModule(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Error = t.useCases.MountModule(request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleMountFunction(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Error = t.useCases.MountFunction(request.Identifiers.Module, request.Identifiers.Function)
	(*response).Success = (*response).Error == nil
}

//////////////////////////////////////////////////////////////////////////////////////////
////							UnMount Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleUnMountModule(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Error = t.useCases.UnMountModule(request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleUnMountFunction(request *thread2.Request, response **thread2.Response) {

	*response = thread2.NewResponse(thread2.Processor)
	(*response).Error = t.useCases.UnMountFunction(request.Identifiers.Module, request.Identifiers.Function)
	(*response).Success = (*response).Error == nil
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Log Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleLogRun(request *thread2.Request, response **thread2.Response) {

	t.requestStore[request.Nonce] = request
	thread2.AsyncLogToRunner(t.channels.c13, request)
}

//////////////////////////////////////////////////////////////////////////////////////////
////					    ~~~ HANDLE RESPONSES ~~~
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleDatabaseReturnsPipeline(iRequest *thread2.Request, iResponse *thread2.Response) {

	if iResponse.Error != nil {
		delete(t.requestStore, iRequest.Nonce)
		oResponse := thread2.NewResponse(thread2.Processor)
		oResponse.Error = iResponse.Error
		t.sendResponse(iRequest, oResponse)
		return
	}

	metadata, ok := (iRequest.Data).(map[string]string)
	if !ok {
		delete(t.requestStore, iRequest.Nonce)
		oResponse := thread2.NewResponse(thread2.Processor)
		oResponse.Error = thread2.BadRequestType
		t.sendResponse(iRequest, oResponse)
		return
	}

	var p *processor2.Processor
	oResponse := thread2.NewResponse(thread2.Processor)
	p, oResponse.Error = t.useCases.FindCandidateProcessor(iResponse)
	if oResponse.Error == nil {
		t.requestStore[iRequest.Nonce] = iRequest
		thread2.AsyncCreateRun(t.channels.c13, iRequest.Nonce, iRequest.Identifiers.Namespace,
			iRequest.Identifiers.Module, iRequest.Identifiers.Pipeline, p.Id, metadata)
	} else {
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
	}
}

func (t *Thread) handleRunnerRespondsToCreateRun(iRequest *thread2.Request, iResponse *thread2.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread2.NewResponse(thread2.Processor)
	if iResponse.Error != nil {
		oResponse.Error = iResponse.Error
	} else {
		oResponse.Data, oResponse.Error = t.useCases.UpdateProcessorAfterRunStarted(iResponse)
	}
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToUpdateRun(iRequest *thread2.Request, iResponse *thread2.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread2.NewResponse(thread2.Processor)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToDeleteRun(iRequest *thread2.Request, iResponse *thread2.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread2.NewResponse(thread2.Processor)
	// TODO : fix?
	//oResponse.Error = t.useCases.CheckIfRunStopped(iResponse)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToLog(iRequest *thread2.Request, iResponse *thread2.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread2.NewResponse(thread2.Processor)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToGet(iRequest *thread2.Request, iResponse *thread2.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread2.NewResponse(thread2.Processor)
	oResponse.Data = iResponse.Data
	oResponse.Error = iResponse.Error

	t.sendResponse(iRequest, oResponse)
}
