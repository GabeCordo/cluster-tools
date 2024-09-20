package interfaces

import "sync"

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
	Create SupervisorEvent = "create"
	Start                  = "start"
	Cancel                 = "cancel"
	Error                  = "error"
)

type Supervisor struct {
	Id     uint64           `json:"id"`
	Status SupervisorStatus `json:"status,omitempty"`

	Processor string `json:"processor,omitempty"`
	Module    string `json:"module,omitempty"`
	Cluster   string `json:"cluster,omitempty"`

	Config     Config      `json:"config,omitempty"`
	Statistics *Statistics `json:"statistics"`

	mutex sync.RWMutex
}

type Log struct {
	Id      uint64
	Level   HTTPLogLevel
	Message string
}
