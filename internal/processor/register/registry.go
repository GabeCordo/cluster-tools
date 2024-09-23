package register

import (
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/supervisor"
)

func NewRegistry(moduleName, clusterName string, clusterImplementation cluster.Cluster) *Registry {
	registry := new(Registry)

	registry.supervisors = make(map[uint64]*supervisor.Supervisor)
	registry.idReference = 0

	registry.module = moduleName
	registry.cluster = clusterName

	return registry
}
