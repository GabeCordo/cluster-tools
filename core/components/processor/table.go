package processor

import (
	"fmt"
	"github.com/GabeCordo/etl-light/module"
	processor_i "github.com/GabeCordo/etl-light/processor"
)

func (table *Table) AddProcessor(cfg *processor_i.Config) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	for _, processor := range table.processors {
		if (processor.Host == cfg.Host) && (processor.Port == cfg.Port) {
			return AlreadyExists
		}
	}

	processor := newProcessor(cfg.Host, cfg.Port)
	table.processors = append(table.processors, processor)

	return nil
}

func (table *Table) GetProcessors() []Processor {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	processors := make([]Processor, len(table.processors))

	// make a copy of the processor list
	for idx, processor := range table.processors {
		processors[idx] = *processor
	}

	return processors
}

func (table *Table) RegisterModule(processorName string, config *module.Config) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	var processorInstance *Processor
	for _, instance := range table.processors {
		name := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
		if processorName == name {
			processorInstance = instance
			break
		}
	}

	/* the operator can only register a module to an existing processor endpoint */
	if processorInstance == nil {
		return DoesNotExist
	}

	/* the operator can not assign the same module to a processor endpoint */
	for _, module := range processorInstance.Modules {
		if module == config.Name {
			return ModuleAlreadyRegistered
		}
	}

	/* add the module name to the provisioner for reference */
	processorInstance.Modules = append(processorInstance.Modules, config.Name)

	var moduleInstance *Module

	/* if the module already exists we should try to re-use the existing module allocation */
	if instance, found := table.modules[config.Name]; found {

		// TODO : support different module versions
		if instance.data.Version != config.Version {
			return ModuleVersionClash
		}

		// TODO : support different contacts based on versions
		if (instance.data.Contact.Name != config.Contact.Name) ||
			(instance.data.Contact.Email != config.Contact.Email) {
			return ModuleContactClash
		}

		moduleInstance = instance
	} else {
		moduleInstance = newModule(config.Name, config.Version, config.Contact)
	}

	for _, export := range config.Exports {

		/* does the cluster association already exist in the module? */
		/* Note: this can be the case if the module already existed */
		if clusterInstance, found := moduleInstance.Get(export.Cluster); found {
			clusterInstance.Add(processorInstance)
			continue
		}

		/* if the cluster doesn't exist this is the first time we will have the record */
		moduleInstance.add(export.Cluster)
		clusterInstance, _ := moduleInstance.Get(export.Cluster)

		/* associate the processor as one of the executors for this cluster */
		clusterInstance.Add(processorInstance)

		/* if this is the first time creating this cluster, we should follow the default
		   mount request outlined by the module config
		*/
		if export.StaticMount {
			clusterInstance.Mount()
		}
	}

	table.modules[config.Name] = moduleInstance

	return nil
}

func (table *Table) Get(name string) (instance *Module, found bool) {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	instance, found = table.modules[name]
	return instance, found
}

func (table *Table) Registered() []ModuleData {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	modules := make([]ModuleData, len(table.modules))

	idx := 0
	for _, instance := range table.modules {
		modules[idx] = instance.data
		idx++
	}

	return modules
}
