package processor

import (
	"sync"
	"time"
)

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
}

func newProcessor(host string, port int) *Processor {
	processor := new(Processor)

	processor.Host = host
	processor.Port = port
	processor.LastUpdate = time.Now()

	return processor
}

type Cluster struct {
	Name    string
	Mounted bool

	Processors []*Processor
	mutex      sync.Mutex
}

func newCluster(name string) *Cluster {
	cluster := new(Cluster)

	cluster.Name = name
	cluster.Mounted = false
	cluster.Processors = make([]*Processor, 0)

	return cluster
}

type Module struct {
	Name    string
	Mounted bool

	Clusters map[string]*Cluster
	mutex    sync.RWMutex
}

func newModule(name string) *Module {
	module := new(Module)

	module.Name = name
	module.Mounted = false
	module.Clusters = make(map[string]*Cluster)

	return module
}

type Table struct {
	Processors      []*Processor
	NumOfProcessors uint8

	Modules map[string]*Module
	mutex   sync.RWMutex
}

func NewTable() *Table {

	table := new(Table)

	table.Processors = make([]*Processor, 0)
	table.NumOfProcessors = 0
	table.Modules = make(map[string]*Module)

	return table
}
