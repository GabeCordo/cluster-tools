package runner

import (
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {

}

func (t *Thread) Start() {

	var iReq *thread.Request
	var iRsp *thread.Response
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-t.channels.c13:
			{
				oRsp = t.HandleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c14 <- oRsp
				}
			}
		case iRsp = <-t.channels.c10:
			{
				var ok bool
				iReq, ok = t.requestStore[iRsp.Nonce]
				if ok {
					t.handleResponse(iReq, iRsp)
				}
			}
		case iRsp = <-t.channels.c16:
			{
				var ok bool
				iReq, ok = t.requestStore[iRsp.Nonce]
				if ok {
					t.handleResponse(iReq, iRsp)
				}
			}
		case <-t.channels.close:
			{
				// shutting down the runner thread
				break
			}
		}
		oRsp = nil
	}
}

func (t *Thread) sendResponse(request *thread.Request, response *thread.Response) {

	thread.CopyMetadata(request, response)
	response.Success = response.Error == nil

	switch request.Source {
	case thread.Processor:
		{
			t.channels.c14 <- response
		}
	default:
		{
			// NOP
		}
	}
}

// HandleRequest
// handles synchronous and asynchronous requests to the Thread.
//
// The runner thread may send a thread.Request when it needs more information
// to process the current thread.Request. Once it receives additional information
// it will send a thread.Response to the original callee.
func (t *Thread) HandleRequest(request *thread.Request) (response *thread.Response) {

	var err error

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.handleGetRun(request, &response)
				}
			default:
				{
					err = thread.BadRequestType
				}
			}
		}
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.handleCreateRun(request, &response)
				}
			default:
				{
					err = thread.BadRequestType
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
					err = thread.BadRequestType
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
					err = thread.BadRequestType
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.handleDeleteRun(request, &response)
				}
			default:
				{
					err = thread.BadRequestType
				}
			}
		}
	default:
		{
			err = thread.BadRequestType
		}
	}

	if err != nil {
		response = thread.NewResponse(thread.Runner)
		response.Error = err
	}

	return response
}

func (t *Thread) handleResponse(iRequest *thread.Request, iResponse *thread.Response) {

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
			case thread.CreateAction:
				{
					switch iResponse.Type {
					case thread.StatisticRecord:
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
	case thread.Socket:
		{
			switch iResponse.Action {
			case thread.CreateAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							t.handleSocketCreatesRun(iRequest, iResponse)
						}
					default:
						{
							// NOP
						}
					}
				}
			case thread.DeleteAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							t.handleSocketDeletesRun(iRequest, iResponse)
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
	case thread.Messenger:
		switch iResponse.Type {
		case thread.RunRecord:
			{
				t.handleMessengerAckLog(iRequest, iResponse)
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

func (t *Thread) TearDown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
