package processor

import (
	"github.com/GabeCordo/ScalingFunctions"
	"sync"
)

type ModuleData struct {
	Name    string
	Version string
	Contact ScalingFunctions.ContactIR
	Mounted bool
}

type Module struct {
	data ModuleData

	functions map[string]*Function
	mutex     sync.RWMutex
}

func newModule(name string, version string, contact ...ScalingFunctions.ContactIR) *Module {
	module := new(Module)

	module.data.Name = name
	module.data.Version = version

	for _, c := range contact {
		module.data.Contact = c
	}

	module.data.Mounted = false
	module.functions = make(map[string]*Function)

	return module
}

type ModuleFunction struct {
	Name        string   `yaml:"name" json:"name"`
	StaticMount bool     `yaml:"static_mount,omitempty" json:"static_mount,omitempty"`
	Parameters  []string `yaml:"parameters" json:"params"`
	Returns     []string `yaml:"returns" json:"returns"`
}

type ModuleContact struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

type ModuleConfig struct {
	Name        string           `yaml:"name" json:"name"`
	Version     string           `yaml:"version" json:"version"`
	StaticMount bool             `yaml:"static_mount,omitempty" json:"static_mount,omitempty"`
	Contact     ModuleContact    `yaml:"contact,omitempty" json:"contact,omitempty"`
	Exports     []ModuleFunction `yaml:"exports" json:"functions"`
}

func (module *Module) addFunction(builder *ScalingFunctions.FunctionIR) (success bool) {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.functions[builder.Identifier]; found {
		return false
	}

	module.functions[builder.Identifier] = newFunction(builder)
	return true
}

func (module *Module) IsMounted() bool {
	return module.data.Mounted
}

func (module *Module) Mount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	module.data.Mounted = true
}

func (module *Module) Unmount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	module.data.Mounted = false
}

func (module *Module) GetData() ModuleData {

	return module.data
}

func (module *Module) GetFunction(name string) (instance *Function, found bool) {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	instance, found = module.functions[name]
	return instance, found
}

func (module *Module) Registered() []FunctionData {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	clusters := make([]FunctionData, len(module.functions))

	idx := 0
	for _, functionInstance := range module.functions {
		clusters[idx] = functionInstance.GetData()
		idx++
	}

	return clusters
}
