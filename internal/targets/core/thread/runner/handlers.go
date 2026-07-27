package runner

import (
	"errors"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/component/message"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/component/message/log"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/run"
	thread2 "github.com/GabeCordo/FunctionScheduler/internal/targets/core/thread"
)

//////////////////////////////////////////////////////////////////////////////////////////
////					    ~~~ HANDLE REQUESTS ~~~
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleGetRun(request *thread2.Request, response **thread2.Response) {

	rr := t.useCases.GetRuns(request.Identifiers.Namespace,
		request.Identifiers.Pipeline, request.Identifiers.Supervisor,
		request.Metadata.Maximum, request.Metadata.Offset)

	*response = thread2.NewResponse(thread2.Runner)
	(*response).Success = true
	(*response).Data = rr
}

func (t *Thread) handleCountRuns(request *thread2.Request, response **thread2.Response) {

	rr := t.useCases.CountRuns(
		request.Identifiers.Namespace, request.Identifiers.Pipeline,
	)

	*response = thread2.NewResponse(thread2.Runner)
	(*response).Success = true
	(*response).Data = rr
}

func (t *Thread) handleCreateRun(request *thread2.Request, response **thread2.Response) {

	// the callee triggering the run sends a pipeline identifier
	// the runner shall look-up the pipeline record to send to the processor
	_, ok := (request.Data).(map[string]string)
	if !ok {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = thread2.BadRequestType
		(*response).Success = false
		return
	}

	// send a request to the Database thread for the pipeline
	t.requestStore[request.Nonce] = request
	thread2.AsyncGetPipelineFromDatabase(t.channels.c15, t.logger, request)
}

func (t *Thread) handleUpdateRun(request *thread2.Request, response **thread2.Response) {

	tmp, ok := (request.Data).(*run.Run)
	if !ok {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = thread2.BadRequestType
		(*response).Success = false
		return
	}

	r, err := t.useCases.GetRun(tmp.Id)
	if err != nil {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = err
		(*response).Success = false
		return
	}

	r.SetStatus(tmp.Status)
	err = r.SetStatistic(tmp.Statistics)
	if err != nil {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = err
		(*response).Success = false
		return
	}

	status := r.GetStatus()
	if (status == run.Completed) || (status == run.Crashed) || (status == run.Terminated) {
		t.logger.Printf("[proc: %d -> FunctionScheduler][id: %d] runner has completed\n", r.Processor, r.GetId())
		t.requestStore[request.Nonce] = request
		thread2.AsyncCreateStatisticRecordInDatabase(t.channels.c15, t.logger, request, r)
	}
}

func (t *Thread) handleDeleteRun(request *thread2.Request, response **thread2.Response) {

	rr := t.useCases.GetRuns(database.Empty, database.Empty, request.Identifiers.Supervisor, database.Zero, database.Zero)

	if len(rr) < 1 {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = errors.New("no run found with the provided id")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	r := rr[0]
	r.Status = run.Cancelled

	t.requestStore[request.Nonce] = request
	thread2.AsyncDeleteRun(t.channels.c9, t.logger, request, request.Identifiers.Supervisor, r.Processor)
}

func (t *Thread) handleLogRun(request *thread2.Request, response **thread2.Response) {

	l, ok := (request.Data).(*log.Log)
	if !ok {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = errors.New("expected a *log.Log type in the Data field")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	r, err := t.useCases.GetRun(l.Id)
	if err != nil {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = err
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	if !r.IsRunning() {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = errors.New("cannot log on a runner that is not running")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	// TODO : I think we can do better than this, I just want a bullet tracer
	var logType thread2.RequestType
	if l.Priority == message.Fatal {
		logType = thread2.FatalLogRecord
	} else if l.Priority == message.Warning {
		logType = thread2.WarningLogRecord
	} else {
		logType = thread2.DefaultLogRecord
	}

	thread2.AsyncSendLogToMessenger(t.channels.c17, t.logger, request.Nonce,
		r.Namespace, r.Pipeline.Identifier, r.GetId(), logType, l.Message)
}

func (t *Thread) handleStopRun(request *thread2.Request, response **thread2.Response) {

	r, err := t.useCases.GetRun(request.Identifiers.Supervisor)
	if err != nil {
		*response = thread2.NewResponse(thread2.Runner)
		(*response).Error = errors.New("expected a *log.Log type in the Data field")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	r.Status = run.Cancelled

	thread2.AsyncSendStopToMessenger(t.channels.c9, t.logger, request, request.Identifiers.Supervisor, r.Processor)
}

//////////////////////////////////////////////////////////////////////////////////////////
////					    ~~~ HANDLE RESPONSES ~~~
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleDatabaseReturnsPipeline(iRequest *thread2.Request, iResponse *thread2.Response) {

	var id uint64 = 0                    // set to a value >0 when no err
	var cfg *ScalingFunctions.PipelineIR // set to a valid value when no err
	var err error                        // indicates we could not create a new record

	if !iResponse.Success {
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = iResponse.Error
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	pipelineConfigs, ok := iResponse.Data.([]*ScalingFunctions.PipelineIR)
	if !ok {
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = errors.New("expected response to be []pipeline.Data")
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	if len(pipelineConfigs) < 1 {
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = errors.New("expected response to be []pipeline.Data of length at least 1")
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	cfg = pipelineConfigs[0]

	var startedBy run.StartedBy
	if iRequest.StartedBy == thread2.HttpClient {
		startedBy = run.Operator
	} else if iRequest.StartedBy == thread2.Processor {
		startedBy = run.Processor
	} else if iRequest.StartedBy == thread2.Scheduler {
		startedBy = run.Scheduler
	} else {
		startedBy = run.Unknown
	}

	id, err = t.useCases.CreateRun(iRequest.Identifiers.Namespace,
		iRequest.Identifiers.Pipeline, iRequest.Identifiers.Processor, cfg, startedBy)

	if err != nil {
		// the runner shall inform the iRequest source that the thread was
		// unable to provision a new run record
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = err
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	// send a request to the processor to start a run with the (id, cfg) pair
	metadata, ok := (iRequest.Data).(map[string]string)
	if !ok {
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = thread2.BadRequestType
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	thread2.AsyncSendRunToSocket(t.channels.c9, t.logger, iRequest, id, cfg, metadata)
}

func (t *Thread) handleDatabaseCreatesStatistic(iRequest *thread2.Request, iResponse *thread2.Response) {

	if iResponse.Error == nil {
		thread2.AsyncCloseMessengerForRun(t.channels.c17, t.logger, iRequest)
	} else {
		// the database failed to create a statistic record for the run
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = iResponse.Error
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
	}
}

func (t *Thread) handleSocketCreatesRun(iRequest *thread2.Request, iResponse *thread2.Response) {

	id, ok := (iResponse.Data).(uint64)
	if !ok {
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = errors.New("the socket did not return a supervisor id")
		oResponse.Success = false
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	r, err := t.useCases.GetRun(id)
	if err != nil {
		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = err
		oResponse.Success = false
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	if iResponse.Error != nil {
		t.logger.Print(iResponse.Error.Error())
		t.logger.Printf("[FunctionScheduler -> proc: %d][id: %d] %s\n", iRequest.Identifiers.Processor, r.GetId(), "could not connect to the processor and runner is canceled")
		r.Status = run.Cancelled

		oResponse := thread2.NewResponse(thread2.Runner)
		oResponse.Error = iResponse.Error
		oResponse.Success = false
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	t.logger.Printf("[FunctionScheduler -> proc: %d][id: %d] %s\n", iRequest.Identifiers.Processor, r.GetId(), "connected to processor and runner is active")
	r.Status = run.Active

	oResponse := thread2.NewResponse(thread2.Runner)
	oResponse.Success = true
	oResponse.Data = iResponse.Data
	delete(t.requestStore, iRequest.Nonce)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleSocketDeletesRun(iRequest *thread2.Request, iResponse *thread2.Response) {

	oResponse := thread2.NewResponse(thread2.Runner)
	oResponse.Error = iResponse.Error
	delete(t.requestStore, iRequest.Nonce)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleMessengerAckLog(iRequest *thread2.Request, iResponse *thread2.Response) {

	oResponse := thread2.NewResponse(thread2.Runner)
	oResponse.Error = iResponse.Error
	delete(t.requestStore, iRequest.Nonce)
	t.sendResponse(iRequest, oResponse)
}
