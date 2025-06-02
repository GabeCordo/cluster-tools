package runner

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/message"
	"github.com/GabeCordo/Flock/internal/core/message/log"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
	"strconv"
)

func (t *Thread) getSupervisor(filter *database.Filter) ([]*run.Run, error) {

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

func (t *Thread) createRun(processorId uint64, namespaceName, pipelineName string, metadata map[string]string) (uint64, error) {

	// TODO : change it so that configs are received via pointer over the channel
	mandatory := thread.Mandatory{
		Pipe:          t.channels.C15,
		ResponseTable: t.responseTable.database,
		Timeout:       t.config.Timeout,
	}
	conf, found := thread.GetPipelineFromDatabase(mandatory, namespaceName, pipelineName)
	if !found {
		return 0, errors.New("no pipeline with that identifier exists")
	}

	filter := database.Filter{
		Processor: processorId,
		Namespace: namespaceName,
	}
	result, _ := t.registry.Create(filter, &conf)

	id := result.(uint64)
	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(id, 10)})

	if len(results) != 1 {
		return 0, errors.New("failed to create a runner")
	}
	sup := (results[0]).(*run.Run)

	// TODO : need to support sending the received metadata

	runRequest := run.Request{
		Id:        id,
		Namespace: namespaceName,
		Config:    &conf,
		Metadata:  metadata,
	}

	socketRequest := thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.RunRecord,
		Identifiers: thread.RequestIdentifiers{Processor: processorId},
		Data:        runRequest,
		Nonce:       rand.Uint32(),
	}
	t.channels.C9 <- socketRequest

	rsp, timedOut := multithreaded.SendAndWait(t.responseTable.socket, socketRequest.Nonce, t.config.Timeout)
	if timedOut {
		t.Logger.Printf("[flock -> %s][id: %d] %s\n", processorId, sup.GetId(), "could not connect to the processor and runner is canceled")
		sup.Status = run.Cancelled
		return 0, errors.New("could not send create run to processor")
	}

	socketResponse := rsp.(thread.Response)
	if socketResponse.Error != nil {
		t.Logger.Print(socketResponse.Error.Error())
		t.Logger.Printf("[flock -> proc: %d][id: %d] %s\n", processorId, sup.GetId(), "could not connect to the processor and runner is canceled")
		sup.Status = run.Cancelled
		return 0, socketResponse.Error
	} else {
		t.Logger.Printf("[flock -> proc: %d][id: %d] %s\n", processorId, sup.GetId(), "connected to processor and runner is active")
		sup.Status = run.Active
	}

	return id, socketResponse.Error
}

func (t *Thread) updateRun(instance *run.Run) error {

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(instance.Id, 10)})
	if len(results) != 1 {
		return errors.New("cannot update a runner that does not exist")
	}
	stored := (results[0]).(*run.Run)

	stored.SetStatus(instance.Status)
	stored.SetStatistic(instance.Statistics)

	status := stored.GetStatus()
	if (status == run.Completed) ||
		(status == run.Crashed) ||
		(status == run.Terminated) {
		// TODO : this can probably encapsulate
		request := thread.Request{
			Action: thread.CreateAction,
			Type:   thread.StatisticRecord,
			Identifiers: thread.RequestIdentifiers{
				Namespace: stored.Namespace,
				Pipeline:  stored.Pipeline.Identifier,
			},
			Data:  stored.GetStatistic(),
			Nonce: rand.Uint32(),
		}
		t.channels.C15 <- request

		rsp, didTimeout := multithreaded.SendAndWait(t.responseTable.database, request.Nonce, t.config.Timeout)
		if didTimeout {
			return multithreaded.NoResponseReceived
		}

		// TODO : this can also be encapsulated
		response := (rsp).(thread.Response)
		if !response.Success {
			return errors.New("failed to database statistics of runner")
		}

		msgrRequest := thread.Request{
			Action: thread.CloseAction,
			Identifiers: thread.RequestIdentifiers{
				Namespace:  stored.Namespace,
				Pipeline:   stored.Pipeline.Identifier,
				Supervisor: instance.Id,
			},
			Nonce: rand.Uint32(),
		}
		t.channels.C17 <- msgrRequest
	}

	return nil
}

func (t *Thread) logRun(l *log.Log) error {

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

	request := thread.Request{
		Action: thread.LogAction,
		Type:   logType,
		Identifiers: thread.RequestIdentifiers{
			Namespace:  instance.Namespace,
			Pipeline:   instance.Pipeline.Identifier,
			Supervisor: instance.GetId(),
		},
		Data:  l.Message,
		Nonce: rand.Uint32(),
	}
	t.channels.C17 <- request

	return nil
}

func (t *Thread) stopRun(id uint64) error {

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(id, 10)})
	if len(results) != 1 {
		return errors.New("no run found with the provided id")
	}

	r := (results[0]).(*run.Run)
	r.Status = run.Cancelled

	request := thread.Request{
		Action: thread.DeleteAction,
		Type:   thread.RunRecord,
		Identifiers: thread.RequestIdentifiers{
			Supervisor: id,
			Processor:  r.Processor,
		},
		Nonce: rand.Uint32(),
	}

	t.channels.C9 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.responseTable.socket, request.Nonce, t.config.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	socketResponse, ok := rsp.(thread.Response)
	if !ok {
		return errors.New("the thread did not return a thread.Response")
	}

	return socketResponse.Error
}
