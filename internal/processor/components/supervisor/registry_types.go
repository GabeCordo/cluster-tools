package supervisor

import (
	"github.com/GabeCordo/clarence/cluster"
	"sync"
)

type Registry struct {
	module  string
	cluster string

	status         cluster.Status
	implementation cluster.Cluster
	mounted        bool

	supervisors            map[uint64]*Supervisor
	numOfActiveSupervisors uint64

	idReference uint64
	mutex       sync.RWMutex
}

type IdentifierRegistryPair struct {
	Module   string
	Cluster  string
	Registry *Registry
}
