package provisioner

import (
	"github.com/GabeCordo/clarence/wrapper"
	"sync"
)

const (
	DefaultFrameworkModule = "common"
)

type Provisioner struct {
	modules map[string]*wrapper.Module
	mutex   sync.RWMutex
}
