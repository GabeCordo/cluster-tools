package processor

import (
	"errors"
	"github.com/GabeCordo/mango/core/components/processor"
)

func (thread *Thread) getClusters(name string) ([]processor.ClusterData, error) {

	instance, found := GetTableInstance().Get(name)
	if !found {
		// TODO : replace with proper error
		return nil, errors.New("no cluster found with that module name")
	}

	return instance.Registered(), nil
}

func (thread *Thread) mountCluster(moduleName, clusterName string) error {

	moduleInstance, found := GetTableInstance().Get(moduleName)
	if !found {
		return processor.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.Get(clusterName)
	if !found {
		return processor.ClusterDoesNotExist
	}

	clusterInstance.Mount()
	return nil
}

func (thread *Thread) unmountCluster(moduleName, clusterName string) error {

	moduleInstance, found := GetTableInstance().Get(moduleName)
	if !found {
		return processor.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.Get(clusterName)
	if !found {
		return processor.ClusterDoesNotExist
	}

	clusterInstance.Unmount()
	return nil
}
