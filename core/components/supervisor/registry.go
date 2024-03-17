package supervisor

import (
	"github.com/GabeCordo/cluster-tools/core/interfaces"
)

func (registry *Registry) Create(processorName, moduleName, clusterName string, conf *interfaces.Config) (identifier uint64) {

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

func (registry *Registry) GetBy(filter *Filter) []*Supervisor {

	supervisors := make([]*Supervisor, 0)

	if filter == nil {
		return supervisors
	}

	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	// if an id is provided we ignore the module and cluster
	if filter.Id != 0 {

		if supervisor, found := registry.supervisors[filter.Id]; found {
			supervisors = append(supervisors, supervisor)
		}

		return supervisors
	}

	useCluster := filter.UseCluster()
	useModule := filter.UseModule()

	for _, supervisor := range registry.supervisors {

		if useCluster && (supervisor.Cluster == filter.Cluster) && (supervisor.Module == filter.Module) {
			supervisors = append(supervisors, supervisor)
		} else if useModule && (supervisor.Module == filter.Module) {
			supervisors = append(supervisors, supervisor)
		} else if !useCluster && !useModule {
			supervisors = append(supervisors, supervisor)
		}
	}

	return supervisors
}
