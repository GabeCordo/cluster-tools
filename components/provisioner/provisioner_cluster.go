package provisioner

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/supervisor"
)

func NewClusterWrapper(identifier string, implementation cluster.Cluster) *ClusterWrapper {

	clusterWrapper := new(ClusterWrapper)

	clusterWrapper.registry = supervisor.NewRegistry(identifier, implementation)
	clusterWrapper.Mounted = false

	return clusterWrapper
}

func (clusterWrapper *ClusterWrapper) IsMounted() bool {

	return clusterWrapper.Mounted
}

func (clusterWrapper *ClusterWrapper) Mount() *ClusterWrapper {

	clusterWrapper.Mounted = true
	return clusterWrapper
}

func (clusterWrapper *ClusterWrapper) UnMount() *ClusterWrapper {

	clusterWrapper.Mounted = false
	return clusterWrapper
}

func (clusterWrapper *ClusterWrapper) GetClusterImplementation() cluster.Cluster {
	return clusterWrapper.registry.GetClusterImplementation()
}

func (clusterWrapper *ClusterWrapper) FindSupervisor(id uint64) (instance *supervisor.Supervisor, found bool) {

	instance, found = clusterWrapper.registry.GetSupervisor(id)
	return instance, found
}

func (clusterWrapper *ClusterWrapper) CreateSupervisor(config ...cluster.Config) *supervisor.Supervisor {

	return clusterWrapper.registry.CreateSupervisor(config...)
}

func (clusterWrapper *ClusterWrapper) DeleteSupervisor(identifier uint64) (deleted, found bool) {

	deleted, found = clusterWrapper.registry.DeleteSupervisor(identifier)
	return deleted, found
}

func (clusterWrapper *ClusterWrapper) CanDelete() (canDelete bool) {

	canDelete = true
	for _, supervisorInstance := range clusterWrapper.registry.GetSupervisors() {
		if !supervisorInstance.Deletable() {
			canDelete = false
			break
		}
	}

	return canDelete
}
