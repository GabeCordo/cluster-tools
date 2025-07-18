package processor

import (
	"errors"

	component "github.com/GabeCordo/Flock/internal/core/component/processor"
)

func (uc UseCases) GetFunctions(name string) ([]component.FunctionData, error) {

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
		return component.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return component.FunctionDoesNotExist
	}

	clusterInstance.Mount()
	uc.Logger.Printf("[%s] cluster %s was MOUNTED\n", moduleName, clusterName)
	uc.ProcessorTable.Print()

	return nil
}

func (uc UseCases) UnMountFunction(moduleName, clusterName string) error {

	moduleInstance, found := uc.ProcessorTable.GetModule(moduleName)
	if !found {
		return component.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetFunction(clusterName)
	if !found {
		return component.FunctionDoesNotExist
	}

	clusterInstance.Unmount()
	uc.Logger.Printf("[%s] cluster %s was UNMOUNTED\n", moduleName, clusterName)
	uc.ProcessorTable.Print()

	return nil
}
