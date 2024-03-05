package supervisor

import (
	"github.com/GabeCordo/cluster-tools/core/components/messenger"
	"github.com/GabeCordo/cluster-tools/core/interfaces/cluster"
	"sync"
)

type Status string

const (
	Created    Status = "created"
	Active            = "active"
	Crashed           = "crashed"
	Completed         = "completed"
	Terminated        = "terminated" // this is legacy
	Cancelled         = "cancelled"
)

type Event string

const (
	Create   Event = "create"
	Start          = "start"
	Cancel         = "cancel"
	Error          = "error"
	Complete       = "complete"
)

type Log struct {
	Id      uint64                    `json:"id"`
	Level   messenger.MessagePriority `json:"level"`
	Message string                    `json:"message"`
}

type Supervisor struct {
	Id     uint64 `json:"id"`
	Status Status `json:"status,omitempty"`

	Processor string `json:"processor,omitempty"`
	Module    string `json:"module,omitempty"`
	Cluster   string `json:"cluster,omitempty"`

	Config     cluster.Config      `json:"config,omitempty"`
	Statistics *cluster.Statistics `json:"statistics"`

	mutex sync.RWMutex
}

func newSupervisor(id uint64, processorName, moduleName, clusterName string, conf *cluster.Config) *Supervisor {
	supervisor := new(Supervisor)

	supervisor.Status = Created
	supervisor.Id = id
	supervisor.Processor = processorName
	supervisor.Module = moduleName
	supervisor.Cluster = clusterName
	supervisor.Config = *conf // make a copy
	supervisor.Statistics = cluster.NewStatistics()

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
	registry.counter = 1
	return registry
}

type Schedule struct {
	Minute int `json:"minute", yaml:"minute"`
	Hour   int `json:"hour", yaml:"hour"`
	Day    int `json:"day", yaml:"day"`
	Month  int `json:"month", yaml:"month"`
}

type Job struct {
	Cluster  string   `json:"cluster", yaml:"cluster"`
	Config   string   `json:"config", yaml:"config"`
	Schedule Schedule `json:"schedule", yaml:"schedule"`
}

type Scheduler struct {
	jobs  []Job `json:"jobs", yaml:"jobs"`
	mutex sync.RWMutex
}

func NewScheduler() *Scheduler {

	scheduler := new(Scheduler)
	return scheduler
}
