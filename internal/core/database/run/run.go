package run

import (
	"errors"
	"github.com/Sentmint/cluster-tools/internal/core/database/pipeline"
	"github.com/Sentmint/cluster-tools/internal/core/database/statistic"
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

type Run struct {
	Id     uint64 `json:"id"`
	Status Status `json:"status,omitempty"`

	Processor string `json:"processor,omitempty"`
	Namespace string `json:"namespace,omitempty"`

	Pipeline pipeline.Pipeline `json:"pipeline,omitempty"`

	Statistics *statistic.Statistics `json:"statistics"`

	mutex sync.RWMutex
}

func New(id uint64, processorName, namespaceName string, cfg *pipeline.Pipeline) *Run {
	supervisor := new(Run)

	supervisor.Status = Created
	supervisor.Id = id
	supervisor.Processor = processorName
	supervisor.Namespace = namespaceName
	supervisor.Pipeline = *cfg // copy instance
	supervisor.Statistics = statistic.NewStatistics(0, 0)

	return supervisor
}

func (supervisor *Run) Event(event Event) Status {

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

func (supervisor *Run) GetStatus() Status {

	supervisor.mutex.RLock()
	defer supervisor.mutex.RUnlock()

	return supervisor.Status
}

func (supervisor *Run) SetStatus(status Status) {

	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()

	supervisor.Status = status
}

func (supervisor *Run) GetId() uint64 {
	return supervisor.Id
}

func (supervisor *Run) GetPipeline() *pipeline.Pipeline {
	return &supervisor.Pipeline
}

func (supervisor *Run) GetStatistic() *statistic.Statistics {
	return supervisor.Statistics
}

func (supervisor *Run) SetStatistic(statistic *statistic.Statistics) error {
	if statistic == nil {
		return errors.New("statistic is nil")
	}
	supervisor.Statistics = statistic
	return nil
}

func (supervisor *Run) IsRunning() bool {

	supervisor.mutex.RLock()
	defer supervisor.mutex.RUnlock()

	return supervisor.Status == Active
}
