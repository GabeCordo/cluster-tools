package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/api"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/message"
	"github.com/GabeCordo/cluster-tools/internal/message/log"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
	"strconv"
)

func (t *Thread) getSupervisor(filter *database.Filter) ([]*supervisor.Supervisor, error) {

	if filter == nil {
		return nil, errors.New("given nil pointer filter")
	}

	instances := t.registry.Get(*filter)
	supervisors := make([]*supervisor.Supervisor, len(instances))
	for i, instance := range instances {
		supervisors[i] = instance.(*supervisor.Supervisor)
	}

	return supervisors, nil
}

func (t *Thread) createSupervisor(processorName, moduleName, clusterName, configName string, metadata map[string]string) (uint64, error) {

	// TODO : change it so that configs are received via pointer over the channel
	mandatory := thread.Mandatory{t.C15, t.DatabaseResponseTable, t.config.Timeout}
	conf, found := thread.GetConfigFromDatabase(mandatory, moduleName, configName)
	if !found {
		return 0, errors.New("no config with that identifier exists")
	}

	filter := database.Filter{
		Module:    moduleName,
		Cluster:   clusterName,
		Config:    configName,
		Processor: processorName,
	}
	result, _ := t.registry.Create(filter, &conf)

	id := result.(uint64)
	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(id, 10)})

	if len(results) != 1 {
		return 0, errors.New("failed to create a supervisor")
	}
	sup := (results[0]).(*supervisor.Supervisor)

	// TODO : need to support sending the received metadata
	err := api.ProvisionSupervisor(processorName, moduleName, clusterName, id, &conf, metadata)

	if err != nil {
		t.Logger.Print(err.Error())
		t.Logger.Printf("[cluster-tools -> %s][id: %d] %s\n", processorName, sup.GetId(), "could not connect to the processor and supervisor is canceled")
		sup.Status = supervisor.Cancelled
	} else {
		t.Logger.Printf("[cluster-tools -> %s][id: %d] %s\n", processorName, sup.GetId(), "connected to processor and supervisor is active")
		sup.Status = supervisor.Active
	}

	return id, err
}

func (t *Thread) updateSupervisor(instance *supervisor.Supervisor) error {

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(instance.Id, 10)})
	if len(results) != 1 {
		return errors.New("cannot update a supervisor that does not exist")
	}
	stored := (results[0]).(*supervisor.Supervisor)

	stored.SetStatus(instance.Status)
	stored.SetStatistic(instance.Statistics)

	t.Logger.Printf("[id: %d][state: %s] updated supervisor record\n", instance.Id, instance.Status)

	status := stored.GetStatus()
	if (status == supervisor.Completed) ||
		(status == supervisor.Crashed) ||
		(status == supervisor.Terminated) {
		// TODO : this can probably encapsulate
		request := thread.Request{
			Action: thread.CreateAction,
			Type:   thread.StatisticRecord,
			Identifiers: thread.RequestIdentifiers{
				Module:  stored.GetModule(),
				Cluster: stored.GetCluster(),
			},
			Data:  stored.GetStatistic(),
			Nonce: rand.Uint32(),
		}
		t.C15 <- request

		rsp, didTimeout := multithreaded.SendAndWait(t.DatabaseResponseTable, request.Nonce, t.config.Timeout)
		if didTimeout {
			return multithreaded.NoResponseReceived
		}

		// TODO : this can also be encapsulated
		response := (rsp).(thread.Response)
		if !response.Success {
			return errors.New("failed to database statistics of supervisor")
		}

		msgrRequest := thread.Request{
			Action: thread.CloseAction,
			Identifiers: thread.RequestIdentifiers{
				Module:     stored.GetModule(),
				Cluster:    stored.GetCluster(),
				Supervisor: instance.Id,
			},
			Nonce: rand.Uint32(),
		}
		t.C17 <- msgrRequest
	}

	return nil
}

func (t *Thread) logSupervisor(l *log.Log) error {

	results := t.registry.Get(database.Filter{Identifier: strconv.FormatUint(l.Id, 10)})

	if len(results) != 1 {
		return errors.New("supervisor does not exist")
	}

	instance, _ := results[0].(*supervisor.Supervisor)

	if !instance.IsRunning() {
		return errors.New("cannot log on a supervisor that is not running")
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
			Module:     instance.GetModule(),
			Cluster:    instance.GetCluster(),
			Supervisor: instance.GetId(),
		},
		Data:  l.Message,
		Nonce: rand.Uint32(),
	}
	t.C17 <- request

	return nil
}
