package provisioner

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"log"
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

func (moduleWrapper *ModuleWrapper) GetClustersData() map[string]bool {

	mounts := make(map[string]bool)

	for identifier, clusterWrapper := range moduleWrapper.clusters {
		mounts[identifier] = clusterWrapper.Mounted
	}

	return mounts
}

func (moduleWrapper *ModuleWrapper) GetClusters() (clusterWrappers []*ClusterWrapper) {

	clusterWrappers = make([]*ClusterWrapper, 0)

	for _, clusterWrapper := range moduleWrapper.clusters {
		clusterWrappers = append(clusterWrappers, clusterWrapper)
	}

	return clusterWrappers
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
	if !moduleWrapper.MarkForDeletion {
		log.Printf("[provisioner] cannot delete %s - not marked for deletion\n", moduleWrapper.Identifier)
		return false
	}

	canDelete = true
	// look over all the supervisor in a module
	for clusterName, clusterWrapper := range moduleWrapper.clusters {

		if !clusterWrapper.CanDelete() {
			log.Printf("[provisioner][cluster] cannot delete %s\n", clusterName)
			canDelete = false
			break
		}
	}

	return canDelete
}
