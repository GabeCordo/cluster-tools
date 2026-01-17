package processor

import (
	"fmt"
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
	Id         uint64    `json:"Id"`
	RemoteAddr string    `json:"RemoteAddr"`
	Status     Status    `json:"Status"`
	LastUpdate time.Time `json:"LastUpdate"`
	Modules    []string  `json:"Modules"`
	Retries    uint32    `json:"Retries"`
	NumOfRuns  int       `json:"NumOfRuns"`
}

func (processor *Processor) ToString() string {
	return processor.RemoteAddr
}

func (processor *Processor) PrettyPrint() {
	template := "%s (id: %d)\n\t└ Status: %s\n\t└ Last Updated: %s\n\t└ Modules:"
	for _, m := range processor.Modules {
		template += fmt.Sprintf("\n\t\t└ %s", m)
	}
	template += "\n\t└ NumOfRuns: %d"

	output := fmt.Sprintf(template, processor.RemoteAddr, processor.Id,
		processor.Status, processor.LastUpdate, processor.NumOfRuns)
	fmt.Println(output)
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
