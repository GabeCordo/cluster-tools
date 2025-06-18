package provision

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	pipeline2 "github.com/GabeCordo/Flock/internal/processor/component/provision/pipeline"
)

type Provisioner struct {
	Modules map[string]*Module

	Supervisors            map[uint64]*pipeline2.Instance
	numOfActiveSupervisors uint64

	idReference uint64

	mutex sync.RWMutex
}

func New() *Provisioner {
	provisioner := new(Provisioner)

	provisioner.Modules = make(map[string]*Module)
	provisioner.Supervisors = make(map[uint64]*pipeline2.Instance)

	return provisioner
}

func (provisioner *Provisioner) ModuleExists(moduleName string) bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	_, found := provisioner.Modules[moduleName]
	return found
}

func (provisioner *Provisioner) GetModules() []*Module {

	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	modules := make([]*Module, 0)
	for _, moduleWrapper := range provisioner.Modules {
		modules = append(modules, moduleWrapper)
	}

	return modules
}

func (provisioner *Provisioner) GetModule(moduleName string) (instance *Module, found bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	instance, found = provisioner.Modules[moduleName]
	if !found {
		return nil, false
	}

	return instance, found
}

func (provisioner *Provisioner) AddModule(identifier string) (*Module, error) {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if _, found := provisioner.Modules[identifier]; found {
		return nil, errors.New("module already exists")
	}

	mod := new(Module)
	mod.Name = identifier
	mod.Version = "1.0"
	mod.functions = make(map[string]Function)
	provisioner.Modules[identifier] = mod

	return mod, nil
}

func (provisioner *Provisioner) getNextUsableId() uint64 {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if (provisioner.idReference + 1) >= math.MaxUint32 {
		provisioner.idReference = 0
	} else {
		provisioner.idReference++
	}

	return provisioner.idReference
}

func (provisioner *Provisioner) NumberOfActiveSupervisors() uint64 {

	return provisioner.numOfActiveSupervisors
}

func (provisioner *Provisioner) SupervisorExists(id uint64) bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	_, found := provisioner.Supervisors[id]
	return found
}

func (provisioner *Provisioner) CreateSupervisor(namespace string, identifier uint64, metadata map[string]string, core string, pipeline *pipeline.Pipeline) (*pipeline2.Instance, error) {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	functions := make([]any, len(pipeline.Functions))

	for i, f := range pipeline.Functions {

		moduleWrapper, found := provisioner.Modules[f.Module]
		if !found {
			return nil, errors.New("module not found")
		}

		functionWrapper, err := moduleWrapper.GetFunction(f.Identifier)
		if err != nil {
			return nil, errors.New("function not found")
		}

		functions[i] = functionWrapper.Value
	}

	var s *pipeline2.Instance
	s = pipeline2.NewInstance(pipeline, functions, metadata)
	s.Id = identifier

	provisioner.numOfActiveSupervisors++
	provisioner.Supervisors[identifier] = s
	return s, nil
}

func (provisioner *Provisioner) DeleteSupervisor(id uint64) (deleted, found bool) {

	provisioner.mutex.RLock()

	supervisorInstance, found := provisioner.Supervisors[id]
	if !found {
		return false, false
	}

	provisioner.mutex.RUnlock()

	found = true
	if supervisorInstance.Deletable() {
		provisioner.mutex.Lock()
		defer provisioner.mutex.Unlock()
		delete(provisioner.Supervisors, id)
		provisioner.numOfActiveSupervisors--
		deleted = true
	} else {
		deleted = false
	}

	return deleted, found
}

func (provisioner *Provisioner) GetSupervisor(id uint64) (*pipeline2.Instance, bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	if s, found := provisioner.Supervisors[id]; found {
		return s, true
	} else {
		return nil, false
	}
}

func (provisioner *Provisioner) GetSupervisors() []*pipeline2.Instance {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	supervisors := make([]*pipeline2.Instance, 0)

	for _, s := range provisioner.Supervisors {
		supervisors = append(supervisors, s)
	}

	return supervisors
}

func (provisioner *Provisioner) SuspendSupervisors() {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	for _, s := range provisioner.Supervisors {
		fmt.Println("teardown runner")
		s.Teardown()
	}
}
