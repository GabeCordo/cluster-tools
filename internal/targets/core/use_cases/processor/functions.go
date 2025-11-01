package processor

import (
	"errors"
	processor2 "github.com/FortifiedCode/flock/internal/targets/core/component/processor"
)

func (uc UseCases) GetFunctions(name string) ([]processor2.FunctionData, error) {

	instance, found := uc.ProcessorTable.GetModule(name)
	if !found {
		// TODO : replace with proper error
		return nil, errors.New("no cluster found with that module name")
	}

	return instance.Registered(), nil
}

func (uc UseCases) MountFunction(moduleName, clusterName string) error {

	moduleInstance, found := uc.ProcessorTable.GetModule(moduleName)
	if !found {
		return processor2.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return processor2.FunctionDoesNotExist
	}

	clusterInstance.Mount()
	uc.Logger.Printf("[%s] cluster %s was MOUNTED\n", moduleName, clusterName)
	uc.ProcessorTable.Print()

	return nil
}

func (uc UseCases) UnMountFunction(moduleName, clusterName string) error {

	moduleInstance, found := uc.ProcessorTable.GetModule(moduleName)
	if !found {
		return processor2.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return processor2.FunctionDoesNotExist
	}

	clusterInstance.Unmount()
	uc.Logger.Printf("[%s] cluster %s was UNMOUNTED\n", moduleName, clusterName)
	uc.ProcessorTable.Print()

	return nil
}
