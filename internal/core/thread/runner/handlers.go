package runner

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/core/component/message"
	"github.com/FortifiedCode/flock/internal/core/component/message/log"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/core/database/run"
	"github.com/FortifiedCode/flock/internal/core/thread"
	"github.com/FortifiedCode/plover"
)

//////////////////////////////////////////////////////////////////////////////////////////
////					    ~~~ HANDLE REQUESTS ~~~
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleGetRun(request *thread.Request, response **thread.Response) {

	rr := t.useCases.GetRuns(request.Identifiers.Namespace,
		request.Identifiers.Pipeline, request.Identifiers.Supervisor)

	*response = thread.NewResponse(thread.Runner)
	(*response).Success = true
	(*response).Data = rr
}

func (t *Thread) handleCreateRun(request *thread.Request, response **thread.Response) {

	// the callee triggering the run sends a pipeline identifier
	// the runner shall look-up the pipeline record to send to the processor
	_, ok := (request.Data).(map[string]string)
	if !ok {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = thread.BadRequestType
		(*response).Success = false
		return
	}

	// send a request to the Database thread for the pipeline
	t.requestStore[request.Nonce] = request
	thread.AsyncGetPipelineFromDatabase(t.channels.c15, request)
}

func (t *Thread) handleUpdateRun(request *thread.Request, response **thread.Response) {

	tmp, ok := (request.Data).(*run.Run)
	if !ok {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = thread.BadRequestType
		(*response).Success = false
		return
	}

	r, err := t.useCases.GetRun(tmp.Id)
	if err != nil {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = err
		(*response).Success = false
		return
	}

	r.SetStatus(tmp.Status)
	err = r.SetStatistic(tmp.Statistics)
	if err != nil {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = err
		(*response).Success = false
		return
	}

	status := r.GetStatus()
	if (status == run.Completed) || (status == run.Crashed) || (status == run.Terminated) {
		t.logger.Printf("[proc: %d -> flock][id: %d] runner has completed\n", r.Processor, r.GetId())
		t.requestStore[request.Nonce] = request
		thread.AsyncCreateStatisticRecordInDatabase(t.channels.c15, request, r)
	}
}

func (t *Thread) handleDeleteRun(request *thread.Request, response **thread.Response) {

	rr := t.useCases.GetRuns(database.Empty, database.Empty, request.Identifiers.Supervisor)

	if len(rr) < 1 {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = errors.New("no run found with the provided id")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	r := rr[0]
	r.Status = run.Cancelled

	t.requestStore[request.Nonce] = request
	thread.AsyncDeleteRun(t.channels.c9, request, request.Identifiers.Supervisor, r.Processor)
}

func (t *Thread) handleLogRun(request *thread.Request, response **thread.Response) {

	l, ok := (request.Data).(*log.Log)
	if !ok {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = errors.New("expected a *log.Log type in the Data field")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	r, err := t.useCases.GetRun(l.Id)
	if err != nil {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = err
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	if !r.IsRunning() {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = errors.New("cannot log on a runner that is not running")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	// TODO : I think we can do better than this, I just want a bullet tracer
	var logType thread.RequestType
	if l.Priority == message.Fatal {
		logType = thread.FatalLogRecord
	} else if l.Priority == message.Warning {
		logType = thread.WarningLogRecord
	} else {
		logType = thread.DefaultLogRecord
	}

	thread.AsyncSendLogToMessenger(t.channels.c17, request.Nonce,
		r.Namespace, r.Pipeline.Identifier, r.GetId(), logType, l.Message)
}

func (t *Thread) handleStopRun(request *thread.Request, response **thread.Response) {

	r, err := t.useCases.GetRun(request.Identifiers.Supervisor)
	if err != nil {
		*response = thread.NewResponse(thread.Runner)
		(*response).Error = errors.New("expected a *log.Log type in the Data field")
		(*response).Success = false
		t.sendResponse(request, *response)
		return
	}

	r.Status = run.Cancelled

	thread.AsyncSendStopToMessenger(t.channels.c9, request, request.Identifiers.Supervisor, r.Processor)
}

//////////////////////////////////////////////////////////////////////////////////////////
////					    ~~~ HANDLE RESPONSES ~~~
//////////////////////////////////////////////////////////////////////////////////////////

func (t *Thread) handleDatabaseReturnsPipeline(iRequest *thread.Request, iResponse *thread.Response) {

	var id uint64 = 0          // set to a value >0 when no err
	var cfg *plover.PipelineIR // set to a valid value when no err
	var err error              // indicates we could not create a new record

	if !iResponse.Success {
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = iResponse.Error
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	pipelineConfigs, ok := iResponse.Data.([]plover.PipelineIR)
	if !ok {
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = errors.New("expected response to be []pipeline.Pipeline")
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	if len(pipelineConfigs) < 1 {
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = errors.New("expected response to be []pipeline.Pipeline of length at least 1")
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	cfg = &pipelineConfigs[0]

	id, err = t.useCases.CreateRun(iRequest.Identifiers.Namespace,
		iRequest.Identifiers.Pipeline, iRequest.Identifiers.Processor, cfg)

	if err != nil {
		// the runner shall inform the iRequest source that the thread was
		// unable to provision a new run record
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = err
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	// send a request to the processor to start a run with the (id, cfg) pair
	metadata, ok := (iRequest.Data).(map[string]string)
	if !ok {
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = thread.BadRequestType
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	thread.AsyncSendRunToSocket(t.channels.c9, iRequest, id, cfg, metadata)
}

func (t *Thread) handleDatabaseCreatesStatistic(iRequest *thread.Request, iResponse *thread.Response) {

	if iResponse.Error == nil {
		thread.AsyncCloseMessengerForRun(t.channels.c17, iRequest)
	} else {
		// the database failed to create a statistic record for the run
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = iResponse.Error
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
	}
}

func (t *Thread) handleSocketCreatesRun(iRequest *thread.Request, iResponse *thread.Response) {

	id, ok := (iResponse.Data).(uint64)
	if !ok {
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = errors.New("the socket did not return a supervisor id")
		oResponse.Success = false
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	r, err := t.useCases.GetRun(id)
	if err != nil {
		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = err
		oResponse.Success = false
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	if iResponse.Error != nil {
		t.logger.Print(iResponse.Error.Error())
		t.logger.Printf("[flock -> proc: %d][id: %d] %s\n", iRequest.Identifiers.Processor, r.GetId(), "could not connect to the processor and runner is canceled")
		r.Status = run.Cancelled

		oResponse := thread.NewResponse(thread.Runner)
		oResponse.Error = iResponse.Error
		oResponse.Success = false
		delete(t.requestStore, iRequest.Nonce)
		t.sendResponse(iRequest, oResponse)
		return
	}

	t.logger.Printf("[flock -> proc: %d][id: %d] %s\n", iRequest.Identifiers.Processor, r.GetId(), "connected to processor and runner is active")
	r.Status = run.Active

	oResponse := thread.NewResponse(thread.Runner)
	oResponse.Success = true
	oResponse.Data = iResponse.Data
	delete(t.requestStore, iRequest.Nonce)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleSocketDeletesRun(iRequest *thread.Request, iResponse *thread.Response) {

	oResponse := thread.NewResponse(thread.Runner)
	oResponse.Error = iResponse.Error
	delete(t.requestStore, iRequest.Nonce)
	t.sendResponse(iRequest, oResponse)
}

func (t *Thread) handleMessengerAckLog(iRequest *thread.Request, iResponse *thread.Response) {

	oResponse := thread.NewResponse(thread.Runner)
	oResponse.Error = iResponse.Error
	delete(t.requestStore, iRequest.Nonce)
	t.sendResponse(iRequest, oResponse)
}
