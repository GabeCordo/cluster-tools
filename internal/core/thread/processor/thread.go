package processor

import (
	"errors"

	"github.com/GabeCordo/Flock/internal/core/processor"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	var iReq thread.Request
	var oRsp thread.Response

	for {
		select {
		case iReq = <-t.C5:
			{
				t.handleRequest(&iReq, &oRsp)
			}
		case iReq = <-t.C7:
			{
				t.handleRequest(&iReq, &oRsp)
			}
		case iReq = <-t.C18:
			{
				t.handleRequest(&iReq, &oRsp)
			}
		case iRsp := <-t.C12:
			{
				var ok bool
				iReq, ok = t.requestStore[iRsp.Nonce]
				if ok {
					t.handleResponse(&iReq, &iRsp, &oRsp)
				}
			}
		case iRsp := <-t.C14:
			{
				var ok bool
				iReq, ok = t.requestStore[iReq.Nonce]
				if ok {
					t.handleResponse(&iReq, &iRsp, &oRsp)
				}
			}
		case <-t.Interrupt:
			{
				// terminate the thread from processing further
				break
			}
		}
	}
}

func (t *Thread) sendResponse(request *thread.Request, response *thread.Response) {

	response.Nonce = request.Nonce
	response.Success = response.Error == nil

	switch request.Source {
	case thread.HttpClient:
		t.C6 <- *response // TODO: send ptr
	case thread.Socket:
		t.C8 <- *response // TODO: send ptr
	case thread.Scheduler:
		t.C19 <- *response // TODO: send ptr
	default:
		// NOP
	}
}

func (t *Thread) handleRequest(request *thread.Request, response *thread.Response) {

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				{
					response.Data = t.syncGetProcessors()
					t.sendResponse(request, response)
				}
			case thread.ModuleRecord:
				{
					response.Data = t.syncGetModules()
					t.sendResponse(request, response)
				}
			case thread.FunctionRecord:
				{
					response.Data, response.Error = t.syncGetFunctions(request.Identifiers.Module)
					t.sendResponse(request, response)
				}
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = *request
					t.asyncGetRunFromRunner(request)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				{
					cfg := (request.Data).(processor.Config)
					response.Error = t.synchAddProcessor(&cfg)
					t.sendResponse(request, response)
				}
			case thread.ModuleRecord:
				{
					cfg := (request.Data).(processor.ModuleConfig)
					response.Error = t.syncAddModule(request.Identifiers.Processor, &cfg)
					t.sendResponse(request, response)
				}
			case thread.RunRecord:
				{
					// fetch the pipeline from the database is async
					t.requestStore[request.Nonce] = *request
					t.asyncGetPipelineFromDatabase(request)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				{
					cfg := (request.Data).(processor.Config)
					response.Error = t.syncDeleteProcessor(&cfg)
					t.sendResponse(request, response)
				}
			case thread.ModuleRecord:
				{
					response.Error = t.syncDeleteModule(request.Identifiers.Processor, request.Identifiers.Module)
					t.sendResponse(request, response)
				}
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = *request
					t.asyncTellRunnerToStopRun(request)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.UpdateAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = *request
					t.asyncSendUpdateToRunner(request)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.MountAction:
		{
			switch request.Type {
			case thread.ModuleRecord:
				{
					response.Error = t.syncMountModule(request.Identifiers.Module)
					t.sendResponse(request, response)
				}
			case thread.FunctionRecord:
				{
					response.Error = t.syncMountFunction(request.Identifiers.Module, request.Identifiers.Function)
					t.sendResponse(request, response)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.UnMountAction:
		{
			switch request.Type {
			case thread.ModuleRecord:
				{
					response.Error = t.syncUnMountModule(request.Identifiers.Module)
					t.sendResponse(request, response)
				}
			case thread.FunctionRecord:
				{
					response.Error = t.syncUnMountFunction(request.Identifiers.Module, request.Identifiers.Function)
					t.sendResponse(request, response)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.LogAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.requestStore[request.Nonce] = *request
					t.asyncSendLogToRunner(request)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	default:
		{
			response.Error = thread.UnknownRequest
		}
	}

	if errors.Is(response.Error, thread.UnknownRequest) {
		t.sendResponse(request, response)
	}
}

func (t *Thread) handleResponse(iRequest *thread.Request, iResponse *thread.Response, oResponse *thread.Response) {

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
							oResponse.Error = iResponse.Error
							t.sendResponse(iRequest, oResponse)
							return
						}

						var p *processor.Processor
						p, oResponse.Error = t.syncFindCandidateProcessor(iResponse)
						if oResponse.Error == nil {
							t.requestStore[iRequest.Nonce] = *iRequest
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
