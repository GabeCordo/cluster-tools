package supervisor

import "github.com/GabeCordo/mango/core/interfaces/cluster"

func (registry *Registry) Create(processorName, moduleName, clusterName string, conf *cluster.Config) (identifier uint64) {

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	identifier = registry.counter

	supervisor := newSupervisor(identifier, processorName, moduleName, clusterName, conf)
	registry.supervisors[identifier] = supervisor

	registry.counter++

	return identifier
}

func (registry *Registry) Get(identifier uint64) (supervisor *Supervisor, found bool) {

	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	supervisor, found = registry.supervisors[identifier]
	return supervisor, found
}
