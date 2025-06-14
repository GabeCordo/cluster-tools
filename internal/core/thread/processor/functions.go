package processor

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/processor"
)

func (t *Thread) syncGetFunctions(name string) ([]processor.FunctionData, error) {

	instance, found := t.processorTable.GetModule(name)
	if !found {
		// TODO : replace with proper error
		return nil, errors.New("no cluster found with that module name")
	}

	return instance.Registered(), nil
}

func (t *Thread) syncMountFunction(moduleName, clusterName string) error {

	moduleInstance, found := t.processorTable.GetModule(moduleName)
	if !found {
		return processor.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return processor.FunctionDoesNotExist
	}

	clusterInstance.Mount()
	t.Logger.Printf("[%s] cluster %s was MOUNTED\n", moduleName, clusterName)
	t.processorTable.Print()

	return nil
}

func (t *Thread) syncUnMountFunction(moduleName, clusterName string) error {

	moduleInstance, found := t.processorTable.GetModule(moduleName)
	if !found {
		return processor.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return processor.FunctionDoesNotExist
	}

	clusterInstance.Unmount()
	t.Logger.Printf("[%s] cluster %s was UNMOUNTED\n", moduleName, clusterName)
	t.processorTable.Print()

	return nil
}
