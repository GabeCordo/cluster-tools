package components

import "github.com/GabeCordo/cluster-tools/internal/interfaces"

type ClusterData struct {
	Name    string
	Mounted bool
	Mode    interfaces.EtlMode
}

type Cluster interface {
	Add(processor Processor)
	IsMounted() bool
	IsStream() bool
	Mount()
	Unmount()
	SelectProcessor() Processor
}

type ModuleData struct {
	Name    string
	Version float64
	Contact interfaces.ModuleContact
	Mounted bool
}

type Module interface {
	IsMounted() bool
	Mount()
	Unmount()
	GetCluster(name string) (instance Cluster, found bool)
	Registered() []ClusterData
}

type Processor interface {
	GetHost() string
	GetPort() int
	Retries() uint32
	Retry()
	ResetRetries()
	Print()
	ToString() string
}

type ProcessorTable interface {
	GetProcessors() []Processor
	AddProcessor(cfg *interfaces.ProcessorConfig) error
	RemoveProcessor(cfg *interfaces.ProcessorConfig) error
	GetModule(name string) (instance Module, found bool)
	AddModule(processorName string, config *interfaces.ModuleConfig) error
	RemoveModule(processor, name string) error
	RegisteredModules() []ModuleData
	Print()
}
