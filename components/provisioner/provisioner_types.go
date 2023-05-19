package provisioner

import (
	"github.com/GabeCordo/etl/components/supervisor"
	"sync"
)

type Provisioner struct {
	Registries map[string]*supervisor.Registry `json:"functions"`
	mutex      sync.RWMutex
}
