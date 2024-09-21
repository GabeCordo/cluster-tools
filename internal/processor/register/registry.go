package registry

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/cluster"
	cluster2 "github.com/GabeCordo/cluster-tools/internal/processor/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/supervisor"
	"math"
)

func NewRegistry(moduleName, clusterName string, clusterImplementation cluster.Cluster) *Registry {
	registry := new(Registry)

	registry.supervisors = make(map[uint64]*supervisor.Supervisor)
	registry.idReference = 0

	registry.module = moduleName
	registry.cluster = clusterName
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

func (registry *Registry) NumberOfActiveSupervisors() uint64 {

	return registry.numOfActiveSupervisors
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

func (registry *Registry) CreateSupervisor(identifier uint64, metadata map[string]string, core string, standalone bool, config ...*cluster.Config) *supervisor.Supervisor {

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	helper := cluster2.NewHelper(core, registry.module, registry.cluster, identifier, standalone)

	var s *supervisor.Supervisor
	if len(config) > 0 {
		s = supervisor.NewCustomSupervisor(registry.implementation, config[0], metadata, helper)
	} else {
		s = supervisor.NewSupervisor(registry.implementation, metadata, helper)
	}
	s.Id = identifier

	registry.numOfActiveSupervisors++
	registry.supervisors[identifier] = s
	return s
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
		registry.numOfActiveSupervisors--
		deleted = true
	} else {
		deleted = false
	}

	return deleted, found
}

func (registry *Registry) GetSupervisor(id uint64) (*supervisor.Supervisor, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	if supervisor, found := registry.supervisors[id]; found {
		return supervisor, true
	} else {
		return nil, false
	}
}

func (registry *Registry) GetSupervisors() []*supervisor.Supervisor {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	supervisors := make([]*supervisor.Supervisor, 0)

	for _, supervisor := range registry.supervisors {
		supervisors = append(supervisors, supervisor)
	}

	return supervisors
}

func (registry *Registry) SuspendSupervisors() {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	for _, supervisor := range registry.supervisors {
		fmt.Println("teardown supervisor")
		supervisor.Teardown()
	}
}

func (registry *Registry) GetClusterImplementation() cluster.Cluster {

	return registry.implementation
}

func (registry *Registry) Event(event supervisor.Event) *Registry {
	switch registry.status {
	case cluster.UnMounted:
		{
			switch event {
			case cluster2.Mount:
				{
					registry.mounted = true
					registry.status = cluster.Mounted
				}
			case cluster2.Delete:
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
			case cluster2.Delete:
				{
					registry.mounted = false
					registry.status = cluster.MarkedForDeletion
				}
			}
		}
	}

	return registry
}
