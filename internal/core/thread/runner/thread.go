package runner

import (
	"errors"
	"strconv"

	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/message/log"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {
	t.accepting = true
}

func (t *Thread) Start() {

	var iReq thread.Request
	var iRsp thread.Response
	var oRsp thread.Response

	for {
		select {
		case iReq = <-t.channels.C13:
			{
				t.handleRequest(&iReq, &oRsp)
			}
		case iRsp = <-t.channels.C10:
			{
				var ok bool
				iReq, ok = t.requestStore[iRsp.Nonce]
				if ok {
					t.handleResponse(&iReq, &iRsp, &oRsp)
				}
			}
		case iRsp = <-t.channels.C16:
			{
				var ok bool
				iReq, ok = t.requestStore[iRsp.Nonce]
				if ok {
					t.handleResponse(&iReq, &iRsp, &oRsp)
				}
			}
		case <-t.Interrupt:
			{
				// Terminate the thread
				break
			}
		}
	}
}

func (t *Thread) sendResponse(request *thread.Request, response *thread.Response) {

	switch request.Source {
	case thread.Processor:
		{
			t.channels.C14 <- *response // todo: pass pointer
		}
	default:
		{
			// NOP
		}
	}
}

func (t *Thread) handleRequest(request *thread.Request, response *thread.Response) {

	response.Source = thread.Runner
	response.Action = request.Action
	response.Type = request.Type
	response.Nonce = request.Nonce

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					f := &database.Filter{
						Namespace:  request.Identifiers.Namespace,
						Pipeline:   request.Identifiers.Pipeline,
						Identifier: strconv.FormatUint(request.Identifiers.Supervisor, 10),
					}
					response.Data, response.Error = t.syncGetSupervisor(f)
				}
			default:
				{
					t.Logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					// the callee triggering the run sends a pipeline identifier
					// the runner shall look-up the pipeline record to send to the processor
					_, ok := (request.Data).(map[string]string)
					if ok {
						t.requestStore[request.Nonce] = *request
						t.asyncGetPipelineFromDatabase(request)
					} else {
						response.Error = errors.New("RunnerCreate expected a map[string]string data type")
						t.sendResponse(request, response)
					}
				}
			default:
				{
					t.Logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	case thread.UpdateAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					var r *run.Run
					r, response.Error = t.syncUpdateRun(request)
					if response.Error == nil {
						status := r.GetStatus()
						if (status == run.Completed) || (status == run.Crashed) || (status == run.Terminated) {
							t.Logger.Printf("run completed %d\n", r.GetId())
							t.requestStore[request.Nonce] = *request
							t.asyncCreateStatisticRecordInDatabase(request, r)
						}
					} else {
						t.sendResponse(request, response)
					}
				}
			default:
				{
					t.Logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	case thread.LogAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					l := (request.Data).(*log.Log)
					response.Error = t.logRun(l)
				}
			default:
				{
					t.Logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					response.Error = t.asyncStopRun(request.Identifiers.Supervisor)
					if response.Error != nil {
						t.sendResponse(request, response)
					} else {
						t.requestStore[request.Nonce] = *request
					}
				}
			default:
				{
					t.Logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	default:
		{
			t.Logger.Warn(thread.UnknownRequest.Error())
			response.Error = thread.BadRequestType
		}
	}

	response.Success = response.Error == nil
}

func (t *Thread) handleResponse(iRequest *thread.Request, iResponse *thread.Response, oResponse *thread.Response) {
	switch iResponse.Source {
	case thread.Database:
		{
			switch iResponse.Action {
			case thread.GetAction:
				{
					switch iResponse.Type {
					case thread.PipelineRecord:
						{
							var id uint64 = 0
							var cfg pipeline.Pipeline
							id, cfg, oResponse.Error = t.syncCreateNewRunRecord(iRequest, iResponse)
							if oResponse.Error == nil {
								t.asyncSendRunToSocket(iRequest, id, &cfg)
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
				}
			case thread.CreateAction:
				{
					switch iResponse.Type {
					case thread.StatisticRecord:
						{
							if iResponse.Error == nil {
								t.asyncCloseMessengerForRun(iRequest)
							} else {
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
