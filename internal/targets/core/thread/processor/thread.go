package processor

import (
	"github.com/GabeCordo/FunctionScheduler/internal/flags"
	"github.com/GabeCordo/FunctionScheduler/internal/shared/terminal"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/thread"
)

func (t *Thread) Setup() {

	t.logger.SetColour(terminal.Orange)
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
		case <-t.channels.close:
			{
				// shutting down the processor thread
				break
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

	if flags.DEBUG {
		t.logger.Printf("Received %s", request.ToString())
	}

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.ProcessorRecord:
				{
					t.handleGetProcessor(request, &response)
				}
			case thread.ModuleRecord:
				{
					t.handleGetModule(request, &response)
				}
			case thread.FunctionRecord:
				{
					t.handleGetFunctions(request, &response)
				}
			case thread.RunRecord:
				{
					t.handleGetRun(request, &response)
				}
			default:
				{
					err = thread.UnknownRequest
				}
			}
		}
	case thread.CountAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.handleCountRuns(request, &response)
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
					t.handleCreateProcessor(request, &response)
				}
			case thread.ModuleRecord:
				{
					t.handleCreateModule(request, &response)
				}
			case thread.RunRecord:
				{
					t.handleCreateRun(request, &response)
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
					t.handleDeleteProcessor(request, &response)
				}
			case thread.ModuleRecord:
				{
					t.handleDeleteModule(request, &response)
				}
			case thread.RunRecord:
				{
					t.handleDeleteRun(request, &response)
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
					t.handleUpdateRun(request, &response)
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
					t.handleMountModule(request, &response)
				}
			case thread.FunctionRecord:
				{
					t.handleMountFunction(request, &response)
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
					t.handleUnMountModule(request, &response)
				}
			case thread.FunctionRecord:
				{
					t.handleUnMountFunction(request, &response)
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
					t.handleLogRun(request, &response)
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

	if flags.DEBUG {
		t.logger.Printf("Received %s", iResponse.ToString())
	}

	switch iResponse.Source {
	case thread.Database:
		{
			switch iResponse.Action {
			case thread.GetAction:
				{
					switch iResponse.Type {
					case thread.PipelineRecord:
						{
							t.handleDatabaseReturnsPipeline(iRequest, iResponse)
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
	case thread.Runner:
		{
			switch iResponse.Action {
			case thread.CreateAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							t.handleRunnerRespondsToCreateRun(iRequest, iResponse)
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
						t.handleRunnerRespondsToUpdateRun(iRequest, iResponse)
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
							t.handleRunnerRespondsToDeleteRun(iRequest, iResponse)
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
							t.handleRunnerRespondsToLog(iRequest, iResponse)
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
							t.handleRunnerRespondsToGet(iRequest, iResponse)
						}
					default:
						{
							// NOP
						}
					}
				}
			case thread.CountAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							t.handleRunnerRespondsToCount(iRequest, iResponse)
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

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
