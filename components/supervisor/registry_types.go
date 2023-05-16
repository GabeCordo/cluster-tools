package supervisor

import (
	"sync"
)

type Registry struct {
	Supervisors map[uint64]*Supervisor

	idReference uint64
	mutex       sync.RWMutex
}

type IdentifierRegistryPair struct {
	Identifier string
	Registry   *Registry
}
