package provisioner

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
)

func NewModuleWrapper() *ModuleWrapper {

	moduleWrapper := new(ModuleWrapper)

	moduleWrapper.clusters = make(map[string]*ClusterWrapper)
	moduleWrapper.mounted = false

	return moduleWrapper
}

func (moduleWrapper *ModuleWrapper) IsMounted() bool {

	return moduleWrapper.mounted
}

func (moduleWrapper *ModuleWrapper) Mount() *ModuleWrapper {

	moduleWrapper.mounted = true
	return moduleWrapper
}

func (moduleWrapper *ModuleWrapper) UnMount() *ModuleWrapper {

	moduleWrapper.mounted = false
	return moduleWrapper
}

func (moduleWrapper *ModuleWrapper) GetClusters() map[string]bool {

	mounts := make(map[string]bool)

	for identifier, clusterWrapper := range moduleWrapper.clusters {
		mounts[identifier] = clusterWrapper.mounted
	}

	return mounts
}

func (moduleWrapper *ModuleWrapper) GetCluster(clusterName string) (clusterWrapper *ClusterWrapper, found bool) {

	moduleWrapper.mutex.RLock()
	defer moduleWrapper.mutex.RUnlock()

	clusterWrapper, found = moduleWrapper.clusters[clusterName]
	return clusterWrapper, found
}

func (moduleWrapper *ModuleWrapper) AddCluster(clusterName string, implementation cluster.Cluster) (*ClusterWrapper, bool) {

	moduleWrapper.mutex.RLock()

	if _, found := moduleWrapper.clusters[clusterName]; found {
		return nil, false
	}

	moduleWrapper.mutex.RUnlock()

	moduleWrapper.mutex.Lock()
	defer moduleWrapper.mutex.Unlock()

	clusterWrapper := NewClusterWrapper(clusterName, implementation)
	moduleWrapper.clusters[clusterName] = clusterWrapper

	fmt.Printf("registered cluster %s\n", clusterName)

	return clusterWrapper, true
}
