package provisioner

import (
	"github.com/GabeCordo/etl/components/supervisor"
	"sync"
)

const (
	DefaultFrameworkModule = "common"
)

type ClusterWrapper struct {
	registry *supervisor.Registry
	mounted  bool

	mutex sync.RWMutex
}

type ModuleWrapper struct {
	clusters map[string]*ClusterWrapper
	mounted  bool

	mutex sync.RWMutex
}

type Provisioner struct {
	modules map[string]*ModuleWrapper
	mutex   sync.RWMutex
}
