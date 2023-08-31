package processor

import (
	"errors"
	"github.com/GabeCordo/mango/core/components/processor"
	"github.com/GabeCordo/mango/core/interfaces/module"
	"github.com/GabeCordo/mango/core/threads/common"
)

func (thread *Thread) getModules() []processor.ModuleData {

	return GetTableInstance().Registered()
}

func (thread *Thread) addModule(processorName string, cfg *module.Config) error {

	if !cfg.Verify() {
		return errors.New("module config is not valid")
	}

	if err := GetTableInstance().RegisterModule(processorName, cfg); err != nil {
		return err
	}

	// the module will send a default config for every cluster it registers within it
	// this config should be used as the de-facto config unless another is specified by the operator
	// -> send the config for storage in the database thread
	for _, export := range cfg.Exports {
		err := common.StoreConfigInDatabase(thread.C11, thread.DatabaseResponseTable, cfg.Name, export.ToClusterConfig(), thread.config.MaxWaitForResponse)
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

	return GetTableInstance().Remove(processorName, moduleName)
}

func (thread *Thread) mountModule(name string) error {

	instance, found := GetTableInstance().Get(name)
	if !found {
		return processor.ModuleDoesNotExist
	}

	instance.Mount()
	thread.Logger.Printf("the module %s was MOUNTED\n", name)
	GetTableInstance().Print()

	return nil
}

func (thread *Thread) unmountModule(name string) error {

	instance, found := GetTableInstance().Get(name)
	if !found {
		return processor.ModuleDoesNotExist
	}

	instance.Unmount()
	thread.Logger.Printf("the module %s was UNMOUNTED\n", name)
	GetTableInstance().Print()

	return nil
}
