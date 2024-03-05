package supervisor

import "github.com/GabeCordo/cluster-tools/core/components/supervisor"

var registryInstance *supervisor.Registry

func GetRegistryInstance() *supervisor.Registry {

	if registryInstance == nil {
		registryInstance = supervisor.NewRegistry()
	}
	return registryInstance
}
