package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/core/database/statistic"
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

type Supervisor struct {
	Id     uint64 `json:"id"`
	Status Status `json:"status,omitempty"`

	Processor string            `json:"processor,omitempty"`
	Module    string            `json:"module,omitempty"`
	Pipeline  string            `json:"pipeline,omitempty"`
	Config    pipeline.Pipeline `json:"pipeline,omitempty"`

	Statistics *statistic.Statistics `json:"statistics"`

	mutex sync.RWMutex
}

func New(id uint64, processorName, moduleName, pipeName string, cfg *pipeline.Pipeline) *Supervisor {
	supervisor := new(Supervisor)

	supervisor.Status = Created
	supervisor.Id = id
	supervisor.Processor = processorName
	supervisor.Module = moduleName
	supervisor.Pipeline = pipeName
	supervisor.Config = *cfg // copy instance
	supervisor.Statistics = statistic.NewStatistics()

	return supervisor
}

func (supervisor *Supervisor) Event(event Event) Status {

	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()

	switch supervisor.Status {
	case Created:
		switch event {
		case Cancel:
			supervisor.Status = Cancelled
		case Start:
			supervisor.Status = Start
		}
	case Active:
		switch event {
		case Complete:
			supervisor.Status = Completed
		case Error:
			supervisor.Status = Crashed
		}
	}

	return supervisor.Status
}

func (supervisor *Supervisor) GetStatus() Status {

	supervisor.mutex.RLock()
	defer supervisor.mutex.RUnlock()

	return supervisor.Status
}

func (supervisor *Supervisor) SetStatus(status Status) {

	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()

	supervisor.Status = status
}

func (supervisor *Supervisor) GetId() uint64 {
	return supervisor.Id
}

func (supervisor *Supervisor) GetModule() string {
	return supervisor.Module
}

func (supervisor *Supervisor) GetPipeline() string {
	return supervisor.Pipeline
}

func (supervisor *Supervisor) GetStatistic() *statistic.Statistics {
	return supervisor.Statistics
}

func (supervisor *Supervisor) SetStatistic(statistic *statistic.Statistics) error {
	if statistic == nil {
		return errors.New("statistic is nil")
	}
	supervisor.Statistics = statistic
	return nil
}

func (supervisor *Supervisor) IsRunning() bool {

	supervisor.mutex.RLock()
	defer supervisor.mutex.RUnlock()

	return supervisor.Status == Active
}
