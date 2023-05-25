package supervisor

import (
	"github.com/GabeCordo/etl/components/cluster"
	"math"
)

func NewRegistry(clusterName string, clusterImplementation cluster.Cluster) *Registry {
	registry := new(Registry)

	registry.supervisors = make(map[uint64]*Supervisor)
	registry.idReference = 0

	registry.identifier = clusterName
	registry.implementation = clusterImplementation
	registry.status = cluster.UnMounted
	registry.mounted = false

	return registry
}

func (registry *Registry) getNextUsableId() uint64 {

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if (registry.idReference + 1) >= math.MaxUint32 {
		registry.idReference = 0
	} else {
		registry.idReference++
	}

	return registry.idReference
}

func (registry *Registry) SupervisorExists(id uint64) bool {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	_, found := registry.supervisors[id]
	return found
}

func (registry *Registry) IsMounted() bool {
	return registry.mounted
}

func (registry *Registry) CreateSupervisor(config ...cluster.Config) *Supervisor {

	id := registry.getNextUsableId()

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	var supervisor *Supervisor
	if len(config) > 0 {
		supervisor = NewCustomSupervisor(registry.implementation, config[0])
	} else {
		supervisor = NewSupervisor(registry.identifier, registry.implementation)
	}
	supervisor.Id = id

	registry.supervisors[id] = supervisor
	return supervisor
}

func (registry *Registry) DeleteSupervisor(id uint64) (deleted, found bool) {

	registry.mutex.RLock()

	supervisorInstance, found := registry.supervisors[id]
	if !found {
		return false, false
	}

	registry.mutex.RUnlock()

	found = true
	if supervisorInstance.Deletable() {
		registry.mutex.Lock()
		defer registry.mutex.Unlock()
		delete(registry.supervisors, id)
		deleted = true
	} else {
		deleted = false
	}

	return deleted, found
}

func (registry *Registry) GetSupervisor(id uint64) (*Supervisor, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	if supervisor, found := registry.supervisors[id]; found {
		return supervisor, true
	} else {
		return nil, false
	}
}

func (registry *Registry) GetSupervisors() []*Supervisor {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	supervisors := make([]*Supervisor, 0)

	for _, supervisor := range registry.supervisors {
		supervisors = append(supervisors, supervisor)
	}

	return supervisors
}

func (registry *Registry) GetClusterImplementation() cluster.Cluster {

	return registry.implementation
}

func (registry *Registry) Event(event Event) *Registry {
	switch registry.status {
	case cluster.UnMounted:
		{
			switch event {
			case cluster.Mount:
				{
					registry.mounted = true
					registry.status = cluster.Mounted
				}
			case cluster.Delete:
				{
					registry.mounted = false
					registry.status = cluster.MarkedForDeletion
				}
			}
		}
	case cluster.Mounted:
		{
			switch event {
			case cluster.UnMounted:
				{
					registry.mounted = false
					registry.status = cluster.UnMounted
				}
			case cluster.Delete:
				{
					registry.mounted = false
					registry.status = cluster.MarkedForDeletion
				}
			}
		}
	}

	return registry
}
