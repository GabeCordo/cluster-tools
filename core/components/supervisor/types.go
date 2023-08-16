package supervisor

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"sync"
)

type Status string

const (
	Created   Status = "created"
	Active           = "active"
	Crashed          = "crashed"
	Completed        = "completed"
	Cancelled        = "cancelled"
)

type Event string

const (
	Create   Event = "create"
	Start          = "start"
	Cancel         = "cancel"
	Error          = "error"
	Complete       = "complete"
)

type Supervisor struct {
	Id uint64

	Status  Status
	Module  string
	Cluster string

	Config cluster.Config

	mutex sync.RWMutex
}

func newSupervisor(id uint64, moduleName, clusterName string, conf *cluster.Config) *Supervisor {
	supervisor := new(Supervisor)

	supervisor.Status = Created
	supervisor.Id = id
	supervisor.Module = moduleName
	supervisor.Cluster = clusterName
	supervisor.Config = *conf // make a copy

	return supervisor
}

type Registry struct {
	supervisors map[uint64]*Supervisor

	counter uint64
	mutex   sync.RWMutex
}

func NewRegistry() *Registry {

	registry := new(Registry)
	registry.supervisors = make(map[uint64]*Supervisor)
	return registry
}
