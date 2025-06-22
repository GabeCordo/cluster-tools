package runner

import (
	"errors"
	"strconv"

	"github.com/GabeCordo/Flock/internal/core/component/message"
	"github.com/GabeCordo/Flock/internal/core/component/message/log"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) syncGetSupervisor(filter *database.Filter) ([]*run.Run, error) {

	if filter == nil {
		return nil, errors.New("given nil pointer filter")
	}

	instances := t.registry.Get(*filter)
	supervisors := make([]*run.Run, len(instances))
	for i, instance := range instances {
		supervisors[i] = instance.(*run.Run)
	}

	return supervisors, nil
}

func (t *Thread) asyncGetPipelineFromDatabase(request *thread.Request) {

	databaseRequest := new(thread.Request)
	if databaseRequest == nil {
		panic("failed to allocated thread.Request")
	}

	databaseRequest.Action = thread.GetAction
	databaseRequest.Type = thread.PipelineRecord
	databaseRequest.Identifiers = thread.RequestIdentifiers{
		Namespace: request.Identifiers.Namespace,
		Pipeline:  request.Identifiers.Pipeline,
	}
	databaseRequest.Source = thread.Runner
	databaseRequest.Nonce = request.Nonce

	t.channels.C15 <- databaseRequest
}

func (t *Thread) syncCreateNewRunRecord(request *thread.Request, response *thread.Response) (uint64, pipeline.Pipeline, error) {

	if !response.Success {
		return 0, pipeline.Pipeline{}, response.Error
	}
	pipelineConfig := response.Data.([]pipeline.Pipeline)[0]

	filter := database.Filter{
		Processor: request.Identifiers.Processor,
		Namespace: request.Identifiers.Namespace,
	}
	result, _ := t.registry.Create(filter, &pipelineConfig)

	id := result.(uint64)
	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(id, 10)})

	if len(results) != 1 {
		return 0, pipeline.Pipeline{}, errors.New("failed to create an internal record for the runner")
	}

	return id, pipelineConfig, nil
}

func (t *Thread) asyncSendRunToSocket(request *thread.Request, id uint64, cfg *pipeline.Pipeline) {

	// TODO : need to support sending the received metadata
	metadata := (request.Data).(map[string]string)

	runRequest := run.Request{
		Id:        id,
		Namespace: request.Identifiers.Namespace,
		Config:    cfg,
		Metadata:  metadata,
	}

	socketRequest := new(thread.Request)
	if socketRequest == nil {
		panic("failed to allocated thread.Request")
	}

	socketRequest.Action = thread.CreateAction
	socketRequest.Type = thread.RunRecord
	socketRequest.Identifiers = thread.RequestIdentifiers{
		Processor:  request.Identifiers.Processor,
		Supervisor: id,
	}
	socketRequest.Data = runRequest
	socketRequest.Source = thread.Runner
	socketRequest.Nonce = request.Nonce

	t.channels.C9 <- socketRequest
}

func (t *Thread) syncUpdateRunAfterFirstResponseFromSocket(request *thread.Request, response *thread.Response) error {

	id, ok := (response.Data).(uint64)
	if !ok {
		return errors.New("the socket did not return a supervisor id")
	}

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(id, 10)})
	if len(results) != 1 {
		return errors.New("could not find supervisor")
	}
	sup := results[0].(*run.Run)

	if response.Error != nil {
		t.Logger.Print(response.Error.Error())
		t.Logger.Printf("[flock -> proc: %d][id: %d] %s\n", request.Identifiers.Processor, sup.GetId(), "could not connect to the processor and runner is canceled")
		sup.Status = run.Cancelled
		return response.Error
	} else {
		t.Logger.Printf("[flock -> proc: %d][id: %d] %s\n", request.Identifiers.Processor, sup.GetId(), "connected to processor and runner is active")
		sup.Status = run.Active
	}

	return nil
}

func (t *Thread) syncUpdateRun(request *thread.Request) (*run.Run, error) {

	r, ok := (request.Data).(*run.Run)
	if !ok {
		return nil, errors.New("received the wrong data type for the request")
	}

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(r.Id, 10)})
	if len(results) != 1 {
		return nil, errors.New("cannot update a runner that does not exist")
	}
	stored := (results[0]).(*run.Run)

	stored.SetStatus(r.Status)
	err := stored.SetStatistic(r.Statistics)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (t *Thread) asyncCreateStatisticRecordInDatabase(request *thread.Request, r *run.Run) {

	req := new(thread.Request)
	if req == nil {
		panic("failed to allocated thread.Request")
	}

	req.Action = thread.CreateAction
	req.Type = thread.StatisticRecord
	req.Identifiers = thread.RequestIdentifiers{
		Namespace: request.Identifiers.Namespace,
		Pipeline:  request.Identifiers.Pipeline,
	}
	req.Data = r.GetStatistic()
	req.Source = thread.Runner
	req.Nonce = request.Nonce

	t.channels.C15 <- req
}

func (t *Thread) asyncCloseMessengerForRun(request *thread.Request) {

	msgrRequest := new(thread.Request)
	if msgrRequest == nil {
		panic("failed to allocated thread.Request")
	}

	msgrRequest.Action = thread.CloseAction
	msgrRequest.Identifiers = thread.RequestIdentifiers{
		Namespace:  request.Identifiers.Namespace,
		Pipeline:   request.Identifiers.Pipeline,
		Supervisor: request.Identifiers.Supervisor,
	}
	msgrRequest.Source = thread.Runner
	msgrRequest.Nonce = request.Nonce

	t.channels.C17 <- msgrRequest
}

func (t *Thread) asyncLogRun(request *thread.Request) error {

	l, ok := (request.Data).(*log.Log)
	if !ok {
		return errors.New("expected a *log.Log type in the Data field")
	}

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(l.Id, 10)})

	if len(results) != 1 {
		return errors.New("runner does not exist")
	}

	instance, _ := results[0].(*run.Run)

	if !instance.IsRunning() {
		return errors.New("cannot log on a runner that is not running")
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

	messengerRequest := new(thread.Request)
	if messengerRequest == nil {
		panic("failed to allocated thread.Request")
	}

	messengerRequest.Action = thread.LogAction
	messengerRequest.Type = logType
	messengerRequest.Identifiers = thread.RequestIdentifiers{
		Namespace:  instance.Namespace,
		Pipeline:   instance.Pipeline.Identifier,
		Supervisor: instance.GetId(),
	}
	messengerRequest.Data = l.Message
	messengerRequest.Nonce = request.Nonce

	t.channels.C17 <- messengerRequest

	return nil
}

func (t *Thread) asyncStopRun(request *thread.Request) error {

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(request.Identifiers.Supervisor, 10)})
	if len(results) != 1 {
		return errors.New("no run found with the provided id")
	}

	r := (results[0]).(*run.Run)
	r.Status = run.Cancelled

	socketRequest := new(thread.Request)
	if socketRequest == nil {
		panic("failed to allocated thread.Request")
	}

	socketRequest.Action = thread.DeleteAction
	socketRequest.Type = thread.RunRecord
	socketRequest.Identifiers = thread.RequestIdentifiers{
		Supervisor: request.Identifiers.Supervisor,
		Processor:  r.Processor,
	}
	socketRequest.Nonce = request.Nonce

	t.channels.C9 <- socketRequest

	return nil
}
