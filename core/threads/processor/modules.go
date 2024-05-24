package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/components/processor"
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"math/rand"
)

func (thread *Thread) getModules() []processor.ModuleData {

	return GetTableInstance().RegisteredModules()
}

func (thread *Thread) addModule(processorName string, cfg *interfaces.ModuleConfig) error {

	if !cfg.Verify() {
		return errors.New("module config is not valid")
	}

	if err := GetTableInstance().AddModule(processorName, cfg); err != nil {
		return err
	}

	// the module will send a default config for every cluster it registers within it
	// this config should be used as the de-facto config unless another is specified by the operator
	// -> send the config for storage in the database thread
	for _, export := range cfg.Exports {
		if export.Config.Mode == interfaces.Stream {
			thread.C13 <- common.ThreadRequest{
				Action: common.CreateAction,
				Type:   common.SupervisorRecord,
				Identifiers: common.RequestIdentifiers{
					Processor: processorName,
					Module:    cfg.Name,
					Cluster:   export.Cluster,
					Config:    export.Cluster,
				},
				Caller: common.System,
				Data:   make(map[string]string),
				Nonce:  rand.Uint32(),
			}
		}

		mandatory := common.ThreadMandatory{thread.C11, thread.DatabaseResponseTable, thread.config.Timeout}
		err := common.StoreConfigInDatabase(mandatory, cfg.Name, export.ToClusterConfig())
		if err == nil {
			thread.Logger.Printf("stored new default config for cluster %s in database\n", export.Cluster)
		} else {
			// the config could have already been stored in a previous module register
			// note: configs are not deleted when the processor is disconnected at the moment
			//		-> the idea is we can re-use them s.t. performance can be improved
			thread.Logger.Printf("failed to store default config for cluster %s in database\n", export.Cluster)
		}
	}

	// let the operator have an understanding of the core's state
	// ->	when a processor is added it may change what modules/configs/processors are available to use
	//		and whether they are mounted in the core currently
	GetTableInstance().Print()

	return nil
}

func (thread *Thread) deleteModule(processorName, moduleName string) error {

	return GetTableInstance().RemoveModule(processorName, moduleName)
}

func (thread *Thread) mountModule(name string) error {

	instance, found := GetTableInstance().GetModule(name)
	if !found {
		return processor.ModuleDoesNotExist
	}

	instance.Mount()
	thread.Logger.Printf("the module %s was MOUNTED\n", name)
	GetTableInstance().Print()

	return nil
}

func (thread *Thread) unmountModule(name string) error {

	instance, found := GetTableInstance().GetModule(name)
	if !found {
		return processor.ModuleDoesNotExist
	}

	instance.Unmount()
	thread.Logger.Printf("the module %s was UNMOUNTED\n", name)
	GetTableInstance().Print()

	return nil
}
