package provision

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/processor"
)

type Module struct {
	Name    string
	Version string

	functions map[string]Function

	mutex sync.RWMutex
}

func (module *Module) AddFunction(name string, value any) error {
	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.functions[name]; found {
		return fmt.Errorf("function %s already exists", name)
	}

	module.functions[name] = Function{Name: name, Value: value}
	return nil
}

func (module *Module) Functions() []*Function {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	functions := make([]*Function, 0, len(module.functions))
	for _, function := range module.functions {
		functions = append(functions, &function)
	}

	return functions
}

func (module *Module) GetFunction(name string) (*Function, error) {
	module.mutex.RLock()
	defer module.mutex.RUnlock()

	if f, found := module.functions[name]; found {
		return &f, nil
	} else {
		return nil, fmt.Errorf("function %s not found", name)
	}
}

func (module *Module) ToConfig() processor.ModuleConfig {

	config := processor.ModuleConfig{}

	config.Name = module.Name
	config.Version = module.Version
	config.StaticMount = true
	// TODO : add contact
	config.Exports = make([]processor.ModuleFunction, len(module.functions))

	i := 0
	for _, f := range module.functions {
		config.Exports[i].Name = f.Name
		config.Exports[i].StaticMount = true

		fReflected := reflect.TypeOf(f.Value)

		config.Exports[i].Parameters = make([]string, fReflected.NumIn())
		for j := 0; j < fReflected.NumIn(); j++ {
			config.Exports[i].Parameters[j] = fReflected.In(j).Name()
		}

		config.Exports[i].Returns = make([]string, fReflected.NumOut())
		for j := 0; j < fReflected.NumOut(); j++ {
			config.Exports[i].Returns[j] = fReflected.Out(j).Name()
		}

		i++
	}

	return config
}
