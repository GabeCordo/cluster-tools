package supervisor

import (
	"errors"
	"github.com/GabeCordo/etl/core/components/supervisor"
	"github.com/GabeCordo/etl/core/threads/common"
)

func (thread *Thread) createSupervisor(moduleName, clusterName, configName string) (uint64, error) {

	// TODO : change it so that configs are received via pointer over the channel
	conf, found := common.GetConfigFromDatabase(thread.C7, thread.DatabaseResponseTable, moduleName, configName)
	if !found {
		return 0, errors.New("no config with that identifier exists")
	}

	identifier := GetRegistryInstance().Create(moduleName, clusterName, &conf)

	// TODO : send the request to the cluster server

	return identifier, nil
}

func (thread *Thread) errorSupervisor(identifier uint64) error {

	if instance, found := GetRegistryInstance().Get(identifier); found {
		instance.Event(supervisor.Error)
	} else {
		return errors.New("supervisor does not exist")
	}

	return nil
}

func (thread *Thread) completeSupervisor(identifier uint64) error {

	if instance, found := GetRegistryInstance().Get(identifier); found {
		instance.Event(supervisor.Complete)
		// TODO : send statistics to database
		// TODO : send complete message to the messenger thread
	} else {
		return errors.New("supervisor does not exist")
	}

	return nil
}
