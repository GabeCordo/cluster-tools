package register

import (
	"sync"
)

type Registry struct {
	module  string
	cluster string

	mutex sync.RWMutex
}

type IdentifierRegistryPair struct {
	Module   string
	Cluster  string
	Registry *Registry
}
