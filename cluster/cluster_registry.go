package cluster

import "math"

func NewRegistry() *Registry {
	registry := new(Registry)

	registry.Supervisors = make(map[uint64]*Supervisor)
	registry.idReference = 0

	return registry
}

func (registry Registry) Exists(id uint64) bool {
	_, found := registry.Supervisors[id]
	return found
}

func (registry *Registry) UnRegister(id uint64) bool {
	if _, found := registry.Supervisors[id]; !found {
		return false
	}
	delete(registry.Supervisors, id)
	return true
}

func (registry *Registry) GetNextUsableId() uint64 {
	registry.mutex.Lock()
	registry.mutex.Unlock()

	if (registry.idReference + 1) >= math.MaxUint32 {
		registry.idReference = 0
	} else {
		registry.idReference++
	}

	return registry.idReference
}

func (registry *Registry) Register(supervisor *Supervisor) (uint64, bool) {
	id := registry.GetNextUsableId()

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if registry.Exists(id) {
		return 0, false
	} else {
		supervisor.Id = id
		registry.Supervisors[id] = supervisor
		return id, true
	}
}

func (registry *Registry) GetSupervisor(id uint64) (*Supervisor, bool) {
	if registry.Exists(id) {
		return registry.Supervisors[id], true
	} else {
		return nil, false
	}
}
