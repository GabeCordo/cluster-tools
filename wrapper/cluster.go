package wrapper

import (
	"github.com/GabeCordo/clarence/cluster"
	"github.com/GabeCordo/clarence/internal/components/supervisor"
	"sync"
)

type Cluster struct {
	registry *supervisor.Registry

	Identifier        string          `json:"identifier"`
	Module            string          `json:"module"`
	Mode              cluster.EtlMode `json:"mode"`
	Mounted           bool            `json:"mounted"`
	MarkedForDeletion bool            `json:"marked-for-deletion"`
	DefaultConfig     cluster.Config  `json:"default-config"`

	mutex sync.RWMutex
}

func NewCluster(moduleName, identifier string, mode cluster.EtlMode, implementation cluster.Cluster) *Cluster {

	clusterWrapper := new(Cluster)

	clusterWrapper.registry = supervisor.NewRegistry(moduleName, identifier, implementation)
	clusterWrapper.Identifier = identifier
	clusterWrapper.Module = moduleName
	clusterWrapper.Mode = mode
	clusterWrapper.Mounted = false

	return clusterWrapper
}

func (clusterWrapper *Cluster) IsStream() bool {

	return clusterWrapper.Mode == cluster.Stream
}

func (clusterWrapper *Cluster) IsMounted() bool {

	return clusterWrapper.Mounted
}

func (clusterWrapper *Cluster) Mount() *Cluster {

	clusterWrapper.Mounted = true
	return clusterWrapper
}

func (clusterWrapper *Cluster) UnMount() *Cluster {

	clusterWrapper.Mounted = false
	return clusterWrapper
}

func (clusterWrapper *Cluster) GetClusterImplementation() cluster.Cluster {
	return clusterWrapper.registry.GetClusterImplementation()
}

func (clusterWrapper *Cluster) FindSupervisors() []*supervisor.Supervisor {
	return clusterWrapper.registry.GetSupervisors()
}

func (clusterWrapper *Cluster) FindSupervisor(id uint64) (instance *supervisor.Supervisor, found bool) {

	instance, found = clusterWrapper.registry.GetSupervisor(id)
	return instance, found
}

func (clusterWrapper *Cluster) CreateSupervisor(identifier uint64, metadata map[string]string, core string, standalone bool, config ...*cluster.Config) *supervisor.Supervisor {

	return clusterWrapper.registry.CreateSupervisor(identifier, metadata, core, standalone, config...)
}

func (clusterWrapper *Cluster) DeleteSupervisor(identifier uint64) (deleted, found bool) {

	deleted, found = clusterWrapper.registry.DeleteSupervisor(identifier)
	return deleted, found
}

func (clusterWrapper *Cluster) SuspendSupervisors() {

	clusterWrapper.registry.SuspendSupervisors()
}

func (clusterWrapper *Cluster) CanDelete() (canDelete bool) {

	canDelete = true
	for _, supervisorInstance := range clusterWrapper.registry.GetSupervisors() {
		if !supervisorInstance.Deletable() {
			canDelete = false
			break
		}
	}

	return canDelete
}
