package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/api"
	"github.com/GabeCordo/cluster-tools/core/components/messenger"
	"github.com/GabeCordo/cluster-tools/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

func (thread *Thread) getSupervisor(filter *supervisor.Filter) ([]*supervisor.Supervisor, error) {

	if filter == nil {
		return nil, errors.New("given nil pointer filter")
	}
	instances := GetRegistryInstance().GetBy(filter)
	return instances, nil
}

func (thread *Thread) createSupervisor(processorName, moduleName, clusterName, configName string, metadata map[string]string) (uint64, error) {

	// TODO : change it so that configs are received via pointer over the channel
	mandatory := common.ThreadMandatory{thread.C15, thread.DatabaseResponseTable, thread.config.Timeout}
	conf, found := common.GetConfigFromDatabase(mandatory, moduleName, configName)
	if !found {
		return 0, errors.New("no config with that identifier exists")
	}

	identifier := GetRegistryInstance().Create(processorName, moduleName, clusterName, &conf)
	sup, _ := GetRegistryInstance().Get(identifier)

	// TODO : need to support sending the received metadata
	err := api.ProvisionSupervisor(processorName, moduleName, clusterName, identifier, &conf, metadata)

	if err != nil {
		thread.Logger.Printf("[core -> %s][id: %d] %s\n", processorName, sup.Id, "could not connect to the processor and supervisor is canceled")
		sup.Status = supervisor.Cancelled
	} else {
		thread.Logger.Printf("[core -> %s][id: %d] %s\n", processorName, sup.Id, "connected to processor and supervisor is active")
		sup.Status = supervisor.Active
	}

	return identifier, err
}

func (thread *Thread) updateSupervisor(instance *supervisor.Supervisor) error {

	stored, found := GetRegistryInstance().Get(instance.Id)
	if !found {
		return errors.New("cannot update a supervisor that does not exist")
	}

	stored.Status = instance.Status
	stored.Statistics = instance.Statistics

	thread.Logger.Printf("[id: %d][state: %s] updated supervisor record\n", instance.Id, instance.Status)

	if (stored.Status == supervisor.Completed) ||
		(stored.Status == supervisor.Crashed) ||
		(stored.Status == supervisor.Terminated) {
		// TODO : this can probably encapsulate
		request := common.ThreadRequest{
			Action: common.CreateAction,
			Type:   common.StatisticRecord,
			Identifiers: common.RequestIdentifiers{
				Module:  stored.Module,
				Cluster: stored.Cluster,
			},
			Data:  stored.Statistics,
			Nonce: rand.Uint32(),
		}
		thread.C15 <- request

		rsp, didTimeout := multithreaded.SendAndWait(thread.DatabaseResponseTable, request.Nonce, thread.config.Timeout)
		if didTimeout {
			return multithreaded.NoResponseReceived
		}

		// TODO : this can also be encapsulated
		response := (rsp).(common.ThreadResponse)
		if !response.Success {
			return errors.New("failed to store statistics of supervisor")
		}

		msgrRequest := common.ThreadRequest{
			Action: common.CloseAction,
			Identifiers: common.RequestIdentifiers{
				Module:     stored.Module,
				Cluster:    stored.Cluster,
				Supervisor: instance.Id,
			},
			Nonce: rand.Uint32(),
		}
		thread.C17 <- msgrRequest
	}

	return nil
}

func (thread *Thread) logSupervisor(log *supervisor.Log) error {

	instance, found := GetRegistryInstance().Get(log.Id)

	if !found {
		return errors.New("supervisor does not exist")
	}

	if !instance.IsRunning() {
		return errors.New("cannot log on a supervisor that is not running")
	}

	// TODO : I think we can do better than this, I just want a bullet tracer
	var logType common.RequestType
	if log.Level == messenger.Fatal {
		logType = common.FatalLogRecord
	} else if log.Level == messenger.Warning {
		logType = common.WarningLogRecord
	} else {
		logType = common.DefaultLogRecord
	}

	request := common.ThreadRequest{
		Action: common.LogAction,
		Type:   logType,
		Identifiers: common.RequestIdentifiers{
			Module:     instance.Module,
			Cluster:    instance.Cluster,
			Supervisor: instance.Id,
		},
		Data:  log.Message,
		Nonce: rand.Uint32(),
	}
	thread.C17 <- request

	return nil
}
