package provisioner

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl/framework/components/supervisor"
	"sync"
)

const (
	DefaultFrameworkModule = "common"
)

type ClusterWrapper struct {
	registry *supervisor.Registry

	Identifier        string          `json:"identifier"`
	Mode              cluster.EtlMode `json:"mode"`
	Mounted           bool            `json:"mounted"`
	MarkedForDeletion bool            `json:"marked-for-deletion"`

	mutex sync.RWMutex
}

type ModuleWrapper struct {
	clusters map[string]*ClusterWrapper

	Mounted         bool `json:"mounted"`
	MarkForDeletion bool `json:"mark-for-deletion"`

	Identifier string  `json:"identifier"`
	Version    float64 `json:"version"`
	Contact    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"contact"`

	mutex sync.RWMutex
}

type Provisioner struct {
	modules map[string]*ModuleWrapper
	mutex   sync.RWMutex
}
