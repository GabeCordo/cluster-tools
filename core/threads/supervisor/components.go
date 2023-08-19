package supervisor

import "github.com/GabeCordo/mango-core/core/components/supervisor"

var registryInstance *supervisor.Registry

func GetRegistryInstance() *supervisor.Registry {

	if registryInstance == nil {
		registryInstance = supervisor.NewRegistry()
	}
	return registryInstance
}
