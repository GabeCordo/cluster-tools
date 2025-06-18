package processor

import (
	"errors"

	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
)

func (t *Thread) syncGetFunctions(name string) ([]processor2.FunctionData, error) {

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
		return processor2.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return processor2.FunctionDoesNotExist
	}

	clusterInstance.Mount()
	t.Logger.Printf("[%s] cluster %s was MOUNTED\n", moduleName, clusterName)
	t.processorTable.Print()

	return nil
}

func (t *Thread) syncUnMountFunction(moduleName, clusterName string) error {

	moduleInstance, found := t.processorTable.GetModule(moduleName)
	if !found {
		return processor2.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return processor2.FunctionDoesNotExist
	}

	clusterInstance.Unmount()
	t.Logger.Printf("[%s] cluster %s was UNMOUNTED\n", moduleName, clusterName)
	t.processorTable.Print()

	return nil
}
