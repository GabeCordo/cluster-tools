package provision

import (
	"sync"

	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/statistic"
)

type RunStatus string

const (
	Created    RunStatus = "created"
	Active               = "active"
	Crashed              = "crashed"
	Completed            = "completed"
	Terminated           = "terminated" // this is legacy
	Cancelled            = "cancelled"
)

type SupervisorEvent string

const (
	Create SupervisorEvent = "create"
	Start                  = "start"
	Cancel                 = "cancel"
	Error                  = "error"
)

type Run struct {
	Id     uint64    `json:"id"`
	Status RunStatus `json:"status,omitempty"`

	Processor string `json:"processor,omitempty"`
	Module    string `json:"module,omitempty"`
	Cluster   string `json:"cluster,omitempty"`

	Config     pipeline.Pipeline     `json:"pipeline,omitempty"`
	Statistics *statistic.Statistics `json:"statistics"`

	mutex sync.RWMutex
}

type LogLevel string

const (
	Normal LogLevel = "normal"
	Warn            = "warn"
	Fatal           = "fatal"
)

type Log struct {
	Id      uint64
	Level   LogLevel
	Message string
}
