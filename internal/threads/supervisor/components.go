package supervisor

import "github.com/GabeCordo/cluster-tools/internal/core/components/supervisor"

var registryInstance *supervisor.Registry

func GetRegistryInstance() *supervisor.Registry {

	if registryInstance == nil {
		registryInstance = supervisor.NewRegistry()
	}
	return registryInstance
}
