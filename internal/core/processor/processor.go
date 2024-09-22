package processor

import (
	"fmt"
	"time"
)

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Status string

const (
	Active   Status = "active"
	Suspect         = "suspect"
	Inactive        = "inactive"
)

type Processor struct {
	Host       string
	Port       int
	Status     Status
	LastUpdate time.Time
	Modules    []string
	Retries    uint32
}

func (processor *Processor) ToString() string {
	return fmt.Sprintf("%s:%d", processor.Host, processor.Port)
}

func newProcessor(host string, port int) *Processor {
	processor := new(Processor)

	processor.Host = host
	processor.Port = port
	processor.Status = Active
	processor.LastUpdate = time.Now()
	processor.Modules = make([]string, 0)

	return processor
}
