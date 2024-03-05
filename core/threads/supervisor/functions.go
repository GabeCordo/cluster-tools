package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/api"
	"github.com/GabeCordo/cluster-tools/core/components/messenger"
	"github.com/GabeCordo/cluster-tools/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

func (thread *Thread) getSupervisor(id uint64) (*supervisor.Supervisor, error) {

	instance, found := GetRegistryInstance().Get(id)
	if !found {
		return nil, errors.New("supervisor does not exist")
	}
	return instance, nil
}

func (thread *Thread) createSupervisor(processorName, moduleName, clusterName, configName string, metadata map[string]string) (uint64, error) {

	// TODO : change it so that configs are received via pointer over the channel
	conf, found := common.GetConfigFromDatabase(thread.C15, thread.DatabaseResponseTable, moduleName, configName, thread.config.Timeout)
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
		request := common.DatabaseRequest{
			Action:  common.DatabaseStore,
			Type:    common.SupervisorStatistic,
			Module:  stored.Module,
			Cluster: stored.Cluster,
			Data:    stored.Statistics,
			Nonce:   rand.Uint32(),
		}
		thread.C15 <- request

		rsp, didTimeout := multithreaded.SendAndWait(thread.DatabaseResponseTable, request.Nonce, thread.config.Timeout)
		if didTimeout {
			return multithreaded.NoResponseReceived
		}

		// TODO : this can also be encapsulated
		response := (rsp).(common.DatabaseResponse)
		if !response.Success {
			return errors.New("failed to store statistics of supervisor")
		}

		msgrRequest := common.MessengerRequest{
			Action:     common.MessengerClose,
			Module:     stored.Module,
			Cluster:    stored.Cluster,
			Supervisor: instance.Id,
			Nonce:      rand.Uint32(),
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
	var action common.MessengerAction
	if log.Level == messenger.Fatal {
		action = common.MessengerFatal
	} else if log.Level == messenger.Warning {
		action = common.MessengerWarning
	} else {
		action = common.MessengerLog
	}

	request := common.MessengerRequest{
		Action:     action,
		Module:     instance.Module,
		Cluster:    instance.Cluster,
		Supervisor: instance.Id,
		Message:    log.Message,
		Nonce:      rand.Uint32(),
	}
	thread.C17 <- request

	return nil
}
