package provisioner

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/supervisor"
	"sync"
)

type ClusterWrapper struct {
}

type Provisioner struct {
	RegisteredFunctions  map[string]cluster.Cluster `json:"functions"`
	OperationalFunctions map[string]*cluster.Cluster
	Registries           map[string]*supervisor.Registry `json:"registries"`

	mutex sync.RWMutex
}
