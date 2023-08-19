package processor

import (
	"fmt"
	"github.com/GabeCordo/mango/module"
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
	Modules    []string
}

func (processor Processor) ToString() string {
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

type ClusterData struct {
	Name    string
	Mounted bool
}

type Cluster struct {
	data ClusterData

	processors      []*Processor
	numOfProcessors int
	processorIndex  int

	mutex sync.Mutex
}

func newCluster(name string) *Cluster {
	cluster := new(Cluster)

	cluster.data.Name = name
	cluster.data.Mounted = false
	cluster.processors = make([]*Processor, 0)
	cluster.processorIndex = 0

	return cluster
}

type ModuleData struct {
	Name    string
	Version float64
	Contact module.Contact
	Mounted bool
}

type Module struct {
	data ModuleData

	clusters map[string]*Cluster
	mutex    sync.RWMutex
}

func newModule(name string, version float64, contact ...module.Contact) *Module {
	module := new(Module)

	module.data.Name = name
	module.data.Version = version

	for _, c := range contact {
		module.data.Contact = c
	}

	module.data.Mounted = false
	module.clusters = make(map[string]*Cluster)

	return module
}

type Table struct {
	processors      []*Processor
	NumOfProcessors uint8

	modules map[string]*Module
	mutex   sync.RWMutex
}

func NewTable() *Table {

	table := new(Table)

	table.processors = make([]*Processor, 0)
	table.NumOfProcessors = 0
	table.modules = make(map[string]*Module)

	return table
}
