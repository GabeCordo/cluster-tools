package provisioner

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
)

func NewModuleWrapper() *ModuleWrapper {

	moduleWrapper := new(ModuleWrapper)

	moduleWrapper.clusters = make(map[string]*ClusterWrapper)
	moduleWrapper.Mounted = false

	return moduleWrapper
}

func (moduleWrapper *ModuleWrapper) IsMounted() bool {

	return moduleWrapper.Mounted
}

func (moduleWrapper *ModuleWrapper) Mount() *ModuleWrapper {

	moduleWrapper.Mounted = true
	return moduleWrapper
}

func (moduleWrapper *ModuleWrapper) UnMount() *ModuleWrapper {

	moduleWrapper.Mounted = false
	return moduleWrapper
}

func (moduleWrapper *ModuleWrapper) GetClusters() map[string]bool {

	mounts := make(map[string]bool)

	for identifier, clusterWrapper := range moduleWrapper.clusters {
		mounts[identifier] = clusterWrapper.Mounted
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

func (moduleWrapper *ModuleWrapper) DeleteCluster(identifier string) (deleted, found bool) {

	clusterWrapper, found := moduleWrapper.clusters[identifier]
	if !found {
		return false, false
	}

	if !clusterWrapper.CanDelete() {
		return false, true
	}

	moduleWrapper.mutex.Lock()
	defer moduleWrapper.mutex.Unlock()

	delete(moduleWrapper.clusters, identifier)
	return true, true
}

func (moduleWrapper *ModuleWrapper) CanDelete() (canDelete bool) {

	moduleWrapper.mutex.RLock()
	defer moduleWrapper.mutex.RUnlock()

	// if the module is not marked for deletion, it should not be deleted
	if moduleWrapper.MarkForDeletion {
		return false
	}

	canDelete = true
	// look over all the supervisor in a module
	for _, clusterWrapper := range moduleWrapper.clusters {

		if !clusterWrapper.CanDelete() {
			canDelete = false
			break
		}
	}

	return canDelete
}
