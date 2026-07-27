package processor

import (
	processor2 "github.com/GabeCordo/FunctionScheduler/internal/targets/core/component/processor"
	thread "github.com/GabeCordo/FunctionScheduler/internal/targets/core/thread"
)

//////////////////////////////////////////////////////////////////////////////////////////
////							~~~ HANDLE REQUESTS ~~~
//////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////////////////
////							Getter Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleGetProcessor(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Data = t.useCases.GetProcessors()
	(*response).Success = true
}

func (t *Thread) handleGetModule(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Data = t.useCases.GetModules()
	(*response).Success = true
}

func (t *Thread) handleGetFunctions(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Data, (*response).Error = t.useCases.GetFunctions(request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleGetRun(request *thread.Request, response **thread.Response) {

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
	thread.AsyncGetRun(
		t.channels.c13, t.logger, request.Nonce,
		request.Identifiers.Namespace, request.Identifiers.Pipeline, request.Identifiers.Supervisor,
		request.Metadata.Maximum, request.Metadata.Offset,
	)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Count Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleCountRuns(request *thread.Request, response **thread.Response) {

	t.requestStore[request.Nonce] = request

	thread.AsyncCountRuns(
		t.channels.c13, t.logger, request.Nonce,
		request.Identifiers.Namespace, request.Identifiers.Pipeline,
	)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Create Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleCreateProcessor(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)

	cfg, ok := (request.Data).(processor2.Config)
	if !ok {
		(*response).Success = false
		(*response).Error = thread.BadRequestType
		return
	}

	(*response).Error = t.useCases.AddProcessor(&cfg)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleCreateModule(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)

	cfg, ok := (request.Data).(*ScalingFunctions.ModuleIR)
	if !ok {
		(*response).Success = false
		(*response).Error = thread.BadRequestType
		return
	}

	(*response).Error = t.useCases.AddModule(request.Identifiers.Processor, cfg)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleCreateRun(request *thread.Request, response **thread.Response) {

	// fetch the pipeline from the database is async
	t.requestStore[request.Nonce] = request
	thread.AsyncGetPipeline(t.channels.c11, t.logger, request.Nonce,
		request.Identifiers.Namespace, request.Identifiers.Pipeline)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Delete Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleDeleteProcessor(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)

	cfg, ok := (request.Data).(processor2.Config)
	if !ok {
		(*response).Success = false
		(*response).Error = thread.BadRequestType
		return
	}

	(*response).Error = t.useCases.DeleteProcessor(&cfg)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleDeleteModule(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Error = t.useCases.DeleteModule(request.Identifiers.Processor, request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleDeleteRun(request *thread.Request, response **thread.Response) {

	t.requestStore[request.Nonce] = request
	thread.AsyncStopRunToRunner(t.channels.c13, t.logger, request)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Update Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleUpdateRun(request *thread.Request, response **thread.Response) {

	t.requestStore[request.Nonce] = request
	thread.AsyncUpdateRunToRunner(t.channels.c13, t.logger, request)
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Mount Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleMountModule(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Error = t.useCases.MountModule(request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleMountFunction(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Error = t.useCases.MountFunction(request.Identifiers.Module, request.Identifiers.Function)
	(*response).Success = (*response).Error == nil
}

//////////////////////////////////////////////////////////////////////////////////////////
////							UnMount Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleUnMountModule(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Error = t.useCases.UnMountModule(request.Identifiers.Module)
	(*response).Success = (*response).Error == nil
}

func (t *Thread) handleUnMountFunction(request *thread.Request, response **thread.Response) {

	*response = thread.NewResponse(thread.Processor)
	(*response).Error = t.useCases.UnMountFunction(request.Identifiers.Module, request.Identifiers.Function)
	(*response).Success = (*response).Error == nil
}

//////////////////////////////////////////////////////////////////////////////////////////
////							Log Functions
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleLogRun(request *thread.Request, response **thread.Response) {

	t.requestStore[request.Nonce] = request
	thread.AsyncLogToRunner(t.channels.c13, t.logger, request)
}

//////////////////////////////////////////////////////////////////////////////////////////
////					    ~~~ HANDLE RESPONSES ~~~
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleDatabaseReturnsPipeline(iRequest *thread.Request, iResponse *thread.Response) {

	if iResponse.Error != nil {
		delete(t.requestStore, iRequest.Nonce)
		oResponse := thread.NewResponse(thread.Processor)
		oResponse.Error = iResponse.Error
		t.sendResponse(iRequest, oResponse)
		return
	}

	metadata, ok := (iRequest.Data).(map[string]string)
	if !ok {
		delete(t.requestStore, iRequest.Nonce)
		oResponse := thread.NewResponse(thread.Processor)
		oResponse.Error = thread.BadRequestType
		t.sendResponse(iRequest, oResponse)
		return
	}

	var p *processor2.Processor
	oResponse := thread.NewResponse(thread.Processor)
	p, oResponse.Error = t.useCases.FindCandidateProcessor(iResponse)
	if oResponse.Error == nil {
		t.requestStore[iRequest.Nonce] = iRequest
		thread.AsyncCreateRun(t.channels.c13, t.logger, iRequest.Nonce, iRequest.Identifiers.Namespace,
			iRequest.Identifiers.Module, iRequest.Identifiers.Pipeline, p.Id, metadata, iRequest.Source)
	} else {
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
	}
}

func (t *Thread) handleRunnerRespondsToCount(iRequest *thread.Request, iResponse *thread.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread.NewResponse(thread.Processor)
	if iResponse.Error != nil {
		oResponse.Error = iResponse.Error
	} else {
		oResponse.Error = nil
		oResponse.Data = iResponse.Data
	}
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToCreateRun(iRequest *thread.Request, iResponse *thread.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread.NewResponse(thread.Processor)
	if iResponse.Error != nil {
		oResponse.Error = iResponse.Error
	} else {
		oResponse.Data, oResponse.Error = t.useCases.UpdateProcessorAfterRunStarted(iResponse)
	}
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToUpdateRun(iRequest *thread.Request, iResponse *thread.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread.NewResponse(thread.Processor)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToDeleteRun(iRequest *thread.Request, iResponse *thread.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread.NewResponse(thread.Processor)
	// TODO : fix?
	//oResponse.Error = t.useCases.CheckIfRunStopped(iResponse)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToLog(iRequest *thread.Request, iResponse *thread.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread.NewResponse(thread.Processor)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleRunnerRespondsToGet(iRequest *thread.Request, iResponse *thread.Response) {

	delete(t.requestStore, iRequest.Nonce)
	oResponse := thread.NewResponse(thread.Processor)
	oResponse.Data = iResponse.Data
	oResponse.Error = iResponse.Error

	t.sendResponse(iRequest, oResponse)
}
