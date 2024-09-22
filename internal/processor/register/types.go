package register

import (
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/supervisor"
	"sync"
)

type Registry struct {
	module  string
	cluster string

	status         cluster.Status
	implementation cluster.Cluster
	mounted        bool

	supervisors            map[uint64]*supervisor.Supervisor
	numOfActiveSupervisors uint64

	idReference uint64
	mutex       sync.RWMutex
}

type IdentifierRegistryPair struct {
	Module   string
	Cluster  string
	Registry *Registry
}
