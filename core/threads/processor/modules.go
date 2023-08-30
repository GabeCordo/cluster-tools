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

	// TODO : what should we do with the configs that we are getting?
	for _, export := range cfg.Exports {
		common.StoreConfigInDatabase(thread.C11, thread.DatabaseResponseTable, cfg.Name, export.ToClusterConfig(), thread.config.MaxWaitForResponse)
	}

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
	return nil
}

func (thread *Thread) unmountModule(name string) error {

	instance, found := GetTableInstance().Get(name)
	if !found {
		return processor.ModuleDoesNotExist
	}

	instance.Unmount()
	return nil
}
