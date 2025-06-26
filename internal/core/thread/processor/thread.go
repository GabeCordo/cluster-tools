package processor

import (
	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	var iReq *thread.Request
	var iRsp *thread.Response
	var oRsp *thread.Response

	for {

		select {
		case iReq = <-t.channels.c5:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c6 <- oRsp
				}
			}
		case iReq = <-t.channels.c7:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c8 <- oRsp
				}
			}
		case iReq = <-t.channels.c18:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c19 <- oRsp
				}
			}
		case iRsp = <-t.channels.c12:
			{
				var ok bool
				iReq, ok = t.requestStore[iRsp.Nonce]
				if ok {
					t.handleResponse(iReq, iRsp)
				}
			}
		case iRsp = <-t.channels.c14:
			{
				var ok bool
				iReq, ok = t.requestStore[iReq.Nonce]
				if ok {
					t.handleResponse(iReq, iRsp)
				}
			}
		}

		oRsp = nil
	}
}

func (t *Thread) sendResponse(request *thread.Request, response *thread.Response) {

	response.Nonce = request.Nonce
	response.Source = thread.Processor
	response.Success = response.Error == nil

	switch request.Source {
	case thread.HttpClient:
		t.channels.c6 <- response
	case thread.Socket:
		t.channels.c8 <- response
	case thread.Scheduler:
		t.channels.c19 <- response
	default:
		// NOP
	}
}

func (t *Thread) handleRequest(request *thread.Request) (response *thread.Response) {

	var err error

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Data = t.syncGetProcessors()
				}
			case thread.ModuleRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Data = t.syncGetModules()
				}
			case thread.FunctionRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Data, response.Error = t.syncGetFunctions(request.Identifiers.Module)
				}
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = request
					t.asyncGetRunFromRunner(request)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				{
					cfg := (request.Data).(processor2.Config)
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncAddProcessor(&cfg)
				}
			case thread.ModuleRecord:
				{
					cfg := (request.Data).(processor2.ModuleConfig)
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncAddModule(request.Identifiers.Processor, &cfg)
				}
			case thread.RunRecord:
				{
					// fetch the pipeline from the database is async
					t.requestStore[request.Nonce] = request
					t.asyncGetPipelineFromDatabase(request)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				{
					cfg := (request.Data).(processor2.Config)
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncDeleteProcessor(&cfg)
				}
			case thread.ModuleRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncDeleteModule(request.Identifiers.Processor, request.Identifiers.Module)
				}
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = request
					t.asyncTellRunnerToStopRun(request)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	case thread.UpdateAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = request
					t.asyncSendUpdateToRunner(request)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	case thread.MountAction:
		{
			switch request.Type {
			case thread.ModuleRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncMountModule(request.Identifiers.Module)
				}
			case thread.FunctionRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncMountFunction(request.Identifiers.Module, request.Identifiers.Function)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	case thread.UnMountAction:
		{
			switch request.Type {
			case thread.ModuleRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncUnMountModule(request.Identifiers.Module)
				}
			case thread.FunctionRecord:
				{
					response = thread.NewResponse(thread.Processor)
					response.Error = t.syncUnMountFunction(request.Identifiers.Module, request.Identifiers.Function)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	case thread.LogAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = request
					t.asyncSendLogToRunner(request)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	default:
		{
			err = thread.UnknownRequest
		}
	}

	if err != nil {
		response = thread.NewResponse(thread.Runner)
		response.Error = err
	}

	// legacy support
	if response != nil {
		response.Success = response.Error == nil
	}

	return response
}

func (t *Thread) handleResponse(iRequest *thread.Request, iResponse *thread.Response) {

	switch iResponse.Source {
	case thread.Database:
		{
			switch iResponse.Action {
			case thread.GetAction:
				switch iResponse.Type {
				case thread.PipelineRecord:
					{
						if iResponse.Error != nil {
							delete(t.requestStore, iRequest.Nonce)
							oResponse := thread.NewResponse(thread.Processor)
							oResponse.Error = iResponse.Error
							t.sendResponse(iRequest, oResponse)
							return
						}

						var p *processor2.Processor
						oResponse := thread.NewResponse(thread.Processor)
						p, oResponse.Error = t.syncFindCandidateProcessor(iResponse)
						if oResponse.Error == nil {
							t.requestStore[iRequest.Nonce] = iRequest
							t.asyncSendCreateRunToRunner(p, iRequest)
						} else {
							delete(t.requestStore, iRequest.Nonce)
							t.sendResponse(iRequest, oResponse)
						}
					}
				default:
					{
						// NOP
					}
				}
			default:
				{
					// NOP
				}
			}
		}
	case thread.Runner:
		{
			switch iResponse.Action {
			case thread.CreateAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							delete(t.requestStore, iRequest.Nonce)
							oResponse := thread.NewResponse(thread.Processor)
							if iResponse.Error != nil {
								oResponse.Error = iResponse.Error
							} else {
								oResponse.Data, oResponse.Error = t.syncUpdateProcessorAfterRunStarted(iResponse)
							}
							t.sendResponse(iRequest, oResponse)
						}
					default:
						{
							// NOP
						}
					}
				}
			case thread.UpdateAction:
				switch iResponse.Type {
				case thread.RunRecord:
					{
						delete(t.requestStore, iRequest.Nonce)
						oResponse := thread.NewResponse(thread.Processor)
						t.sendResponse(iRequest, oResponse)
					}
				default:
					{
						// NOP
					}
				}
			case thread.DeleteAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							delete(t.requestStore, iRequest.Nonce)
							oResponse := thread.NewResponse(thread.Processor)
							oResponse.Error = t.syncCheckIfRunStopped(iResponse)
							t.sendResponse(iRequest, oResponse)
						}
					default:
						{
							// NOP
						}
					}
				}
			case thread.LogAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							delete(t.requestStore, iRequest.Nonce)
							oResponse := thread.NewResponse(thread.Processor)
							t.sendResponse(iRequest, oResponse)
						}
					default:
						{
							// NOP
						}
					}
				}
			case thread.GetAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							delete(t.requestStore, iRequest.Nonce)
							oResponse := thread.NewResponse(thread.Processor)
							t.sendResponse(iRequest, oResponse)
						}
					default:
						{
							// NOP
						}
					}
				}
			default:
				{
					// NOP
				}
			}
		}
	default:
		{
			// NOP
		}
	}
}

func (t *Thread) Teardown() {
	t.accepting = false
	t.wg.Wait()
}
