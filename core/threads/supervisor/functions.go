package supervisor

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/mango/core/components/supervisor"
	"github.com/GabeCordo/mango/core/threads/common"
)

func (thread *Thread) createSupervisor(processorName, moduleName, clusterName, configName string) (uint64, error) {

	// TODO : change it so that configs are received via pointer over the channel
	conf, found := common.GetConfigFromDatabase(thread.C15, thread.DatabaseResponseTable, moduleName, configName, thread.config.MaxWaitForResponse)
	if !found {
		return 0, errors.New("no config with that identifier exists")
	}

	identifier := GetRegistryInstance().Create(processorName, moduleName, clusterName, &conf)

	// TODO : send the request to the cluster server
	supervisor, _ := GetRegistryInstance().Get(identifier)
	fmt.Println(supervisor)

	return identifier, nil
}

func (thread *Thread) updateSupervisor(supervisor *supervisor.Supervisor) error {

	return nil
}
