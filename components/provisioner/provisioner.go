package provisioner

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/module"
)

func NewProvisioner() *Provisioner {
	provisioner := new(Provisioner)

	provisioner.modules = make(map[string]*ModuleWrapper)
	provisioner.modules[DefaultFrameworkModule] = NewModuleWrapper()

	return provisioner
}

func (provisioner *Provisioner) ModuleExists(moduleName string) bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	_, found := provisioner.modules[moduleName]
	return found
}

func (provisioner *Provisioner) GetModules() []*ModuleWrapper {

	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	modules := make([]*ModuleWrapper, 0)
	for _, moduleWrapper := range provisioner.modules {
		modules = append(modules, moduleWrapper)
	}

	return modules
}

func (provisioner *Provisioner) GetModule(moduleName string) (instance *ModuleWrapper, found bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	instance, found = provisioner.modules[moduleName]
	if !found {
		return nil, false
	}

	if instance.MarkForDeletion {
		return nil, false
	}

	return instance, found
}

func (provisioner *Provisioner) AddModule(implementation *module.Module) (success bool) {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if implementation == nil {
		return false
	}

	moduleWrapper := NewModuleWrapper()
	provisioner.modules[implementation.Config.Identifier] = moduleWrapper

	moduleWrapper.Version = implementation.Config.Version
	moduleWrapper.Identifier = implementation.Config.Identifier
	moduleWrapper.Contact.Name = implementation.Config.Contact.Name
	moduleWrapper.Contact.Email = implementation.Config.Contact.Email

	for _, export := range implementation.Config.Exports {

		f, err := implementation.Plugin.Lookup(export.Cluster)
		if err != nil {
			fmt.Println(err)
			continue
		}

		clusterImplementation, ok := (f).(cluster.Cluster)
		if !ok {
			fmt.Println(ok)
			continue
		}

		clusterWrapper, _ := moduleWrapper.AddCluster(export.Cluster, clusterImplementation)
		clusterWrapper.Mounted = export.StaticMount
	}

	return true
}

func (provisioner *Provisioner) DeleteModule(identifier string) (deleted, markedForDeletion, found bool) {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if moduleWrapper, found := provisioner.modules[identifier]; found {

		found = true

		moduleWrapper.MarkForDeletion = true
		markedForDeletion = true

		if moduleWrapper.CanDelete() {
			delete(provisioner.modules, identifier)
		}
		deleted = true
	} else {
		found = false
	}

	return deleted, markedForDeletion, found
}
