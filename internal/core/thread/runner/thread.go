package runner

import (
	"strconv"

	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/run"
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
			// Flow: Get Run Record (Step 1)
			switch request.Type {
			case thread.RunRecord:
				{
					f := &database.Filter{
						Namespace:  request.Identifiers.Namespace,
						Pipeline:   request.Identifiers.Pipeline,
						Identifier: strconv.FormatUint(request.Identifiers.Supervisor, 10),
					}

					response = thread.NewResponse(thread.Runner)
					response.Data, response.Error = t.syncGetSupervisor(f)
				}
			default:
				{
					err = thread.BadRequestType
				}
			}
		}
	case thread.CreateAction:
		{
			// Flow: Create Run
			//
			// Step 1 -> Pull the pipeline record from the database.
			switch request.Type {
			case thread.RunRecord:
				{
					// send a request to the Database thread for the pipeline
					t.requestStore[request.Nonce] = request
					err = t.asyncGetPipelineFromDatabase(request)
					if err != nil {
						delete(t.requestStore, request.Nonce)
					}
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
					var r *run.Run
					r, err = t.syncUpdateRun(request)

					if err == nil {
						status := r.GetStatus()
						if (status == run.Completed) || (status == run.Crashed) || (status == run.Terminated) {
							t.Logger.Printf("run completed %d\n", r.GetId())
							t.requestStore[request.Nonce] = request
							t.asyncCreateStatisticRecordInDatabase(request, r)
						}
					}
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
					err = t.asyncLogRun(request)
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
					err = t.asyncStopRun(request)
					if err != nil {
						t.sendResponse(request, response)
					} else {
						t.requestStore[request.Nonce] = request
					}
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
					// Flow: Create Run
					//
					// Step 2 -> Received a response from the database
					switch iResponse.Type {
					case thread.PipelineRecord:
						{
							var id uint64 = 0         // set to a value >0 when no err
							var cfg pipeline.Pipeline // set to a valid value when no err
							var err error             // indicates we could not create a new record

							id, cfg, err = t.syncCreateNewRunRecord(iRequest, iResponse)
							if err != nil {
								delete(t.requestStore, iRequest.Nonce)
								// the runner shall inform the iRequest source that the thread was
								// unable to provision a new run record
								oResponse := thread.NewResponse(thread.Runner)
								oResponse.Error = err
								t.sendResponse(iRequest, oResponse)
							} else {
								// send a request to the processor to start a run with the (id, cfg) pair
								t.asyncSendRunToSocket(iRequest, id, &cfg)
							}
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
							if iResponse.Error == nil {

								t.asyncCloseMessengerForRun(iRequest)
							} else {
								// the database failed to create a statistic record for the run
								oResponse := thread.NewResponse(thread.Runner)
								oResponse.Error = iResponse.Error
								delete(t.requestStore, iRequest.Nonce)
								t.sendResponse(iRequest, oResponse)
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
	case thread.Socket:
		{
			switch iResponse.Action {
			case thread.CreateAction:
				{
					switch iResponse.Type {
					case thread.RunRecord:
						{
							oResponse := thread.NewResponse(thread.Runner)
							oResponse.Error = t.syncUpdateRunAfterFirstResponseFromSocket(iRequest, iResponse)
							oResponse.Data = iResponse.Data
							delete(t.requestStore, iRequest.Nonce)
							t.sendResponse(iRequest, oResponse)
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
							oResponse := thread.NewResponse(thread.Runner)
							oResponse.Error = iResponse.Error
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
	case thread.Messenger:
		switch iResponse.Type {
		case thread.RunRecord:
			{
				oResponse := thread.NewResponse(thread.Runner)
				oResponse.Error = iResponse.Error
				delete(t.requestStore, iRequest.Nonce)
				t.sendResponse(iRequest, oResponse)
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

func (t *Thread) Teardown() {
	t.accepting = false
	t.wg.Wait() // don't tear down until all the requests have been processed
}
