package run

import (
	"errors"
	"sync"
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
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

func FromString(s string) Status {
	switch s {
	case "created":
		return Created
	case "active":
		return Active
	case "crashed":
		return Crashed
	case "completed":
		return Completed
	case "terminated":
		return Terminated
	case "cancelled":
		return Cancelled
	default:
		return Cancelled
	}
}

type StartedBy string

const (
	Operator  StartedBy = "operator"
	Processor           = "processor"
	Scheduler           = "scheduler"
	Unknown             = "unknown"
)

func StartedByFromString(s string) StartedBy {
	switch s {
	case "operator":
		return Operator
	case "processor":
		return Processor
	case "scheduler":
		return Scheduler
	default:
		return Unknown
	}
}

type Event string

const (
	Start    = "start"
	Cancel   = "cancel"
	Error    = "error"
	Complete = "complete"
)

type Request struct {
	Id        uint64                       `json:"id"`
	Namespace string                       `json:"namespace"`
	Config    *ScalingFunctions.PipelineIR `json:"config"`
	Metadata  map[string]string            `json:"data"`
}

type Run struct {
	Id     uint64 `json:"id"`
	Status Status `json:"status,omitempty"`

	Time struct {
		Created     time.Time `json:"created,omitempty"`
		LastUpdated time.Time `json:"last_updated,omitempty"`
		Duration    struct {
			Hours       uint64 `json:"hours"`
			Minutes     uint64 `json:"minutes"`
			Seconds     uint64 `json:"seconds"`
			Millisecond uint64 `json:"milliseconds"`
		} `json:"duration,omitempty"`
	} `json:"time"`

	StartedBy StartedBy `json:"started_by,omitempty"`

	Processor uint64 `json:"processor,omitempty"`
	Namespace string `json:"namespace,omitempty"`

	Pipeline ScalingFunctions.PipelineIR `json:"pipeline,omitempty"`

	Statistics *ScalingFunctions.Statistics `json:"statistics"`

	mutex sync.RWMutex
}

func New(runId, processorId uint64, namespaceName string, startedBy StartedBy, cfg *ScalingFunctions.PipelineIR) *Run {
	run := new(Run)

	run.Status = Created
	run.Id = runId
	run.Processor = processorId
	run.Namespace = namespaceName
	run.StartedBy = startedBy
	run.Pipeline = *cfg // copy instance
	run.Statistics = ScalingFunctions.NewStatistics(0, 0)
	run.Time.Created = time.Now()
	run.Time.LastUpdated = run.Time.Created

	return run
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
	supervisor.Time.LastUpdated = time.Now()
	supervisor.calculateDuration()
}

func (supervisor *Run) GetId() uint64 {
	return supervisor.Id
}

func (supervisor *Run) GetPipeline() *ScalingFunctions.PipelineIR {
	return &supervisor.Pipeline
}

func (supervisor *Run) GetStatistic() *ScalingFunctions.Statistics {
	return supervisor.Statistics
}

func (supervisor *Run) SetStatistic(statistic *ScalingFunctions.Statistics) error {
	if statistic == nil {
		return errors.New("statistic is nil")
	}
	supervisor.Time.LastUpdated = time.Now()
	supervisor.Statistics = statistic
	supervisor.calculateDuration()
	return nil
}

func (supervisor *Run) IsRunning() bool {

	supervisor.mutex.RLock()
	defer supervisor.mutex.RUnlock()

	return supervisor.Status == Active
}

func (supervisor *Run) calculateDuration() {

	duration := supervisor.Time.LastUpdated.Sub(supervisor.Time.Created)
	hours := uint64(duration.Hours())
	supervisor.Time.Duration.Hours = hours

	minutes := uint64(duration.Minutes())
	supervisor.Time.Duration.Minutes = minutes - (hours * 60)

	seconds := uint64(duration.Seconds())
	supervisor.Time.Duration.Seconds = seconds - (minutes * 60)

	if (duration.Milliseconds() > 0) && (duration.Milliseconds() < 100000) {
		milliseconds := uint64(duration.Milliseconds()) // #nosec G115
		supervisor.Time.Duration.Millisecond = milliseconds - (seconds * 1000)
	}
}

type Database interface {
	Get(filter database.Filter) []*Run
	Count(filter database.Filter) uint32
	Create(filter database.Filter, data *ScalingFunctions.PipelineIR, startedBy StartedBy) (uint64, error)
	Print()
}
