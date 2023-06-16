package provisioner

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/module"
	"log"
)

func NewProvisioner() *Provisioner {
	provisioner := new(Provisioner)

	provisioner.modules = make(map[string]*ModuleWrapper)

	defaultFrameworkModule := NewModuleWrapper()
	defaultFrameworkModule.Mount()
	provisioner.modules[DefaultFrameworkModule] = defaultFrameworkModule

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

	fmt.Printf("[provisioner] attempting to add module %s\n", implementation.Config.Identifier)

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if implementation == nil {
		return false
	}

	if _, found := provisioner.modules[implementation.Config.Identifier]; found {
		log.Printf("[provisioner] the module %s already exists - it cannot be added\n", implementation.Config.Identifier)
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

		if clusterWrapper, success := moduleWrapper.AddCluster(export.Cluster, clusterImplementation); success {
			clusterWrapper.Mounted = export.StaticMount
		} else {
			log.Printf("[provisioner] could not add cluster %s\n", export.Cluster)
		}
	}

	return true
}

func (provisioner *Provisioner) DeleteModule(identifier string) (deleted, markedForDeletion, found bool) {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	log.Printf("[provisioner] attempting to delete module %s\n", identifier)

	deleted = false

	if moduleWrapper, foundModule := provisioner.modules[identifier]; foundModule {
		found = true

		provisioner.modules[identifier].MarkForDeletion = true
		markedForDeletion = true

		if moduleWrapper.CanDelete() {
			delete(provisioner.modules, identifier)
			deleted = true
			log.Printf("[provisioner] module %s deleted\n", identifier)
		} else {
			log.Printf("[provisioner] could not delete %s\n", identifier)
		}
	} else {
		log.Printf("[provisioner] could not find %s\n", identifier)
		found = false
	}

	return deleted, markedForDeletion, found
}
