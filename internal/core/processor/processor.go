package processor

import (
	"fmt"
	"time"
)

type Config struct {
	Identifier uint64 `json:"identifier"`
}

type Status string

const (
	Active   Status = "active"
	Suspect         = "suspect"
	Inactive        = "inactive"
)

type Processor struct {
	Id         uint64
	Status     Status
	LastUpdate time.Time
	Modules    []string
	Retries    uint32
	NumOfRuns  int
}

func (processor *Processor) ToString() string {
	return fmt.Sprintf("%d", processor.Id)
}

func newProcessor(id uint64) *Processor {
	processor := new(Processor)

	processor.Id = id
	processor.Status = Active
	processor.LastUpdate = time.Now()
	processor.Modules = make([]string, 0)

	return processor
}
