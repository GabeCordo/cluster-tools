package provisioner

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/supervisor"
)

func NewProvisioner() *Provisioner {
	provisioner := new(Provisioner)

	provisioner.RegisteredFunctions = make(map[string]cluster.Cluster)
	provisioner.OperationalFunctions = make(map[string]*cluster.Cluster)
	provisioner.Registries = make(map[string]*supervisor.Registry)

	return provisioner
}

func (provisioner *Provisioner) Register(function string, cluster cluster.Cluster) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if _, found := provisioner.RegisteredFunctions[function]; found {
		return false
	}

	provisioner.RegisteredFunctions[function] = cluster

	provisioner.Registries[function] = supervisor.NewRegistry()

	return true
}

func (provisioner *Provisioner) UnRegister(function string) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if _, found := provisioner.RegisteredFunctions[function]; !found {
		return false
	}

	delete(provisioner.RegisteredFunctions, function)

	if _, found := provisioner.OperationalFunctions[function]; found {
		delete(provisioner.OperationalFunctions, function)
	}

	return true
}

func (provisioner *Provisioner) Function(identifier string) (cluster.Cluster, *supervisor.Registry, bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	if _, found := provisioner.OperationalFunctions[identifier]; !found {
		return nil, nil, false
	}

	clusterInstance := provisioner.OperationalFunctions[identifier]
	registryInstance := provisioner.Registries[identifier]

	return *clusterInstance, registryInstance, true
}

func (provisioner *Provisioner) Mount(identifier string) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	// if the function has already been mounted and is in an operational state, we don't need to do anything
	if _, found := provisioner.OperationalFunctions[identifier]; found {
		return true
	}

	// if the function is not mounted, but exists as a registered function, mount it so that it can be provisioned
	if clusterInstance, found := provisioner.RegisteredFunctions[identifier]; found {
		provisioner.OperationalFunctions[identifier] = &clusterInstance
	} else {
		return false
	}

	return true
}

func (provisioner *Provisioner) UnMount(identifier string) bool {

	if _, found := provisioner.OperationalFunctions[identifier]; !found {
		return false
	}

	delete(provisioner.OperationalFunctions, identifier)

	return true
}

func (provisioner *Provisioner) IsMounted(identifier string) bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	_, found := provisioner.OperationalFunctions[identifier]
	return found
}

func (provisioner *Provisioner) Mounts() map[string]bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	mounts := make(map[string]bool)
	for identifier := range provisioner.RegisteredFunctions {
		mounts[identifier] = false
	}

	for identifier := range provisioner.OperationalFunctions {
		mounts[identifier] = true
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

	if registryInstance, found := provisioner.Registries[clusterIdentifier]; found {
		return registryInstance, true
	} else {
		return nil, false
	}
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
