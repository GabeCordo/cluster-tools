package supervisor

import (
	"github.com/GabeCordo/etl/components/cluster"
	"sync"
)

type Registry struct {
	identifier string

	status         cluster.Status
	implementation cluster.Cluster
	mounted        bool

	supervisors            map[uint64]*Supervisor
	numOfActiveSupervisors uint64

	idReference uint64
	mutex       sync.RWMutex
}

type IdentifierRegistryPair struct {
	Identifier string
	Registry   *Registry
}
