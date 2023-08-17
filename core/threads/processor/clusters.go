package processor

import "github.com/GabeCordo/etl/core/components/processor"

func (thread *Thread) getClusters(name string) ([]processor.ClusterData, bool) {

	instance, found := GetTableInstance().Get(name)
	if !found {
		return nil, false
	}

	return instance.Registered(), true
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
