package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/processor"
)

func (t *Thread) getClusters(name string) ([]processor.ClusterData, error) {

	instance, found := t.processorTable.GetModule(name)
	if !found {
		// TODO : replace with proper error
		return nil, errors.New("no cluster found with that module name")
	}

	return instance.Registered(), nil
}

func (t *Thread) mountCluster(moduleName, clusterName string) error {

	moduleInstance, found := t.processorTable.GetModule(moduleName)
	if !found {
		return processor.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetCluster(clusterName)
	if !found {
		return processor.ClusterDoesNotExist
	}

	clusterInstance.Mount()
	t.Logger.Printf("[%s] cluster %s was MOUNTED\n", moduleName, clusterName)
	t.processorTable.Print()

	return nil
}

func (t *Thread) unmountCluster(moduleName, clusterName string) error {

	moduleInstance, found := t.processorTable.GetModule(moduleName)
	if !found {
		return processor.ModuleDoesNotExist
	}

	clusterInstance, found := moduleInstance.GetCluster(clusterName)
	if !found {
		return processor.ClusterDoesNotExist
	}

	clusterInstance.Unmount()
	t.Logger.Printf("[%s] cluster %s was UNMOUNTED\n", moduleName, clusterName)
	t.processorTable.Print()

	return nil
}
