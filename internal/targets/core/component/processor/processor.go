package processor

import (
	"time"
)

type Config struct {
	Identifier uint64 `json:"identifier"`
	RemoteAddr string `json:"remote_addr"`
}

type Status string

const (
	Active Status = "active"
)

type Processor struct {
	Id         uint64
	RemoteAddr string
	Status     Status
	LastUpdate time.Time
	Modules    []string
	Retries    uint32
	NumOfRuns  int
}

func (processor *Processor) ToString() string {
	return processor.RemoteAddr
}

func newProcessor(id uint64, remoteAddr string) *Processor {
	processor := new(Processor)

	processor.Id = id
	processor.RemoteAddr = remoteAddr
	processor.Status = Active
	processor.LastUpdate = time.Now()
	processor.Modules = make([]string, 0)

	return processor
}
