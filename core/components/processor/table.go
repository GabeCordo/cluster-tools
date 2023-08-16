package processor

import "github.com/GabeCordo/etl-light/module"

func (table *Table) GetProcessors() []Processor {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	processors := make([]Processor, len(table.Processors))

	// make a copy of the processor list
	for idx, processor := range table.Processors {
		processors[idx] = *processor
	}

	return processors
}

func (table *Table) RegisterModule(config module.Config) {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	// TODO : register the module / clusters depending on the config
}

func (table *Table) Get(name string) (instance *Module, found bool) {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	instance, found = table.Modules[name]
	return instance, found
}

func (table *Table) Registered() map[string]bool {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	modules := make(map[string]bool)

	for identifier, moduleInstance := range table.Modules {
		modules[identifier] = moduleInstance.Mounted
	}

	return modules
}
