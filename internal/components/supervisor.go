package components

import "github.com/GabeCordo/cluster-tools/internal/interfaces"

type SupervisorStatus string

const (
	Created    SupervisorStatus = "created"
	Active                      = "active"
	Crashed                     = "crashed"
	Completed                   = "completed"
	Terminated                  = "terminated" // this is legacy
	Cancelled                   = "cancelled"
)

type SupervisorEvent string

const (
	Create   SupervisorEvent = "create"
	Start                    = "start"
	Cancel                   = "cancel"
	Error                    = "error"
	Complete                 = "complete"
)

type Supervisor interface {
	Event(event SupervisorEvent) SupervisorStatus
	GetStatus() SupervisorStatus
	SetStatus(status SupervisorStatus)
	GetId() uint64
	GetModule() string
	GetCluster() string
	GetStatistic() *interfaces.Statistics
	SetStatistic(*interfaces.Statistics) error
	IsRunning() bool
}

type SupervisorFilter struct {
	Module  string
	Cluster string
	Id      uint64
}

type SupervisorRegistry interface {
	Create(processorName, moduleName, clusterName string, conf *interfaces.Config) (identifier uint64)
	Get(identifier uint64) (supervisor Supervisor, found bool)
	GetBy(filter *SupervisorFilter) []Supervisor
}
