package processor

import (
	"github.com/Sentmint/yule"
	"sync"
)

type Module struct {
	Metadata yule.Module
	mounted  bool

	functions map[string]*Function
	mutex     sync.RWMutex
}

func newModule(name string, version string, contact ...yule.ModuleContact) *Module {
	module := new(Module)

	module.Metadata.Name = name
	module.Metadata.Version = version

	for _, c := range contact {
		module.Metadata.Contact = c
	}

	module.mounted = false
	module.functions = make(map[string]*Function)

	return module
}

func (module *Module) addFunction(builder *yule.ModuleFunction) (success bool) {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.functions[builder.Name]; found {
		return false
	}

	module.functions[builder.Name] = newFunction(builder)
	return true
}

func (module *Module) IsMounted() bool {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	return module.mounted
}

func (module *Module) Mount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	module.mounted = true
}

func (module *Module) Unmount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	module.mounted = false
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
