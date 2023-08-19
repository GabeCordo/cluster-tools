package processor

import (
	"errors"
	"github.com/GabeCordo/mango-core/core/components/processor"
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango/module"
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
		common.StoreConfigInDatabase(thread.C11, thread.DatabaseResponseTable, cfg.Name, export.ToClusterConfig())
	}

	return nil
}

func (thread *Thread) deleteModule(processorName, moduleName string) error {

	// TODO : delete module logic
	return nil
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
