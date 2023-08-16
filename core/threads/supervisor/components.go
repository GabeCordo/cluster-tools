package supervisor

import "github.com/GabeCordo/etl/core/components/supervisor"

var registryInstance *supervisor.Registry

func GetRegistryInstance() *supervisor.Registry {

	if registryInstance == nil {
		registryInstance = supervisor.NewRegistry()
	}
	return registryInstance
}
