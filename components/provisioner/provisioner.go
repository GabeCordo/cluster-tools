package provisioner

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/supervisor"
)

func NewProvisioner() *Provisioner {
	provisioner := new(Provisioner)

	provisioner.Registries = make(map[string]*supervisor.Registry)

	return provisioner
}

func (provisioner *Provisioner) Register(function string, cluster cluster.Cluster) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if _, found := provisioner.Registries[function]; found {
		return false
	}

	provisioner.Registries[function] = supervisor.NewRegistry(function, cluster)

	return true
}

func (provisioner *Provisioner) UnRegister(function string) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if registry, found := provisioner.Registries[function]; found {
		registry.Event(cluster.Delete)
	}

	return true
}

func (provisioner *Provisioner) Function(identifier string) (cluster.Cluster, bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	if registry, found := provisioner.Registries[identifier]; found && registry.IsMounted() {
		return registry.GetClusterImplementation(), true
	} else {
		return nil, false
	}
}

func (provisioner *Provisioner) Mount(identifier string) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	// if the function has already been mounted and is in an operational state, we don't need to do anything
	// if the function is not mounted, but exists as a registered function, mount it so that it can be provisioned
	if registry, found := provisioner.Registries[identifier]; found {
		registry.Event(cluster.Mount)
		return true
	} else {
		return false
	}
}

func (provisioner *Provisioner) UnMount(identifier string) bool {

	if registry, found := provisioner.Registries[identifier]; found {
		registry.Event(cluster.UnMount)
		return true
	} else {
		return false
	}
}

func (provisioner *Provisioner) IsMounted(identifier string) bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	registry, found := provisioner.Registries[identifier]
	return found && registry.IsMounted()
}

func (provisioner *Provisioner) Mounts() map[string]bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	mounts := make(map[string]bool)
	for identifier, registry := range provisioner.Registries {
		mounts[identifier] = registry.IsMounted()
	}

	return mounts
}

func (provisioner *Provisioner) DoesClusterExist(clusterIdentifier string) bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	_, found := provisioner.Registries[clusterIdentifier]
	return found
}

func (provisioner *Provisioner) GetRegistry(clusterIdentifier string) (registryInstance *supervisor.Registry, found bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	registryInstance, found = provisioner.Registries[clusterIdentifier]
	return registryInstance, found
}

func (provisioner *Provisioner) GetRegistries() (registries []supervisor.IdentifierRegistryPair) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	registries = make([]supervisor.IdentifierRegistryPair, 0)

	for identifier, registryInstance := range provisioner.Registries {
		registries = append(registries, supervisor.IdentifierRegistryPair{Identifier: identifier, Registry: registryInstance})
	}

	return registries
}
