package provisioner

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/supervisor"
)

func NewClusterWrapper(identifier string, implementation cluster.Cluster) *ClusterWrapper {

	clusterWrapper := new(ClusterWrapper)

	clusterWrapper.registry = supervisor.NewRegistry(identifier, implementation)
	clusterWrapper.mounted = false

	return clusterWrapper
}

func (clusterWrapper *ClusterWrapper) IsMounted() bool {

	return clusterWrapper.mounted
}

func (clusterWrapper *ClusterWrapper) Mount() *ClusterWrapper {

	clusterWrapper.mounted = true
	return clusterWrapper
}

func (clusterWrapper *ClusterWrapper) UnMount() *ClusterWrapper {

	clusterWrapper.mounted = false
	return clusterWrapper
}

func (clusterWrapper *ClusterWrapper) FindSupervisor(id uint64) (instance *supervisor.Supervisor, found bool) {

	instance, found = clusterWrapper.registry.GetSupervisor(id)
	return instance, found
}

func (clusterWrapper *ClusterWrapper) CreateSupervisor(config ...cluster.Config) *supervisor.Supervisor {

	return clusterWrapper.registry.CreateSupervisor(config...)
}
