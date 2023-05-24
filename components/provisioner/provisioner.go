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

func (provisioner *Provisioner) GetModules() map[string]bool {

	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	modules := make(map[string]bool)
	for identifier, moduleWrapper := range provisioner.modules {
		modules[identifier] = moduleWrapper.mounted
	}

	return modules
}

func (provisioner *Provisioner) GetModule(moduleName string) (instance *ModuleWrapper, found bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	instance, found = provisioner.modules[moduleName]
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
		clusterWrapper.mounted = export.StaticMount
	}

	return true
}
