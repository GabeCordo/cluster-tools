package processor

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
)

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

func (table *Table) AddProcessor(cfg *interfaces.ProcessorConfig) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	for _, processor := range table.processors {
		if (processor.Host == cfg.Host) && (processor.Port == cfg.Port) {
			return AlreadyExists
		}
	}

	processor := newProcessor(cfg.Host, cfg.Port)
	table.processors = append(table.processors, processor)
	table.NumOfProcessors++

	return nil
}

// RemoveProcessor
// this is a REALLY expensive operation that might need to be optimized in the future.
func (table *Table) RemoveProcessor(cfg *interfaces.ProcessorConfig) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	idx := 0
	var instance *Processor = nil
	for idx, instance = range table.processors {
		if (instance.Host == cfg.Host) && (instance.Port == cfg.Port) {
			break
		}
	}

	if instance == nil {
		return errors.New("processor does not exist")
	}

	table.processors = append(table.processors[:idx], table.processors[idx+1:]...)

	// Removing a processor modifies the Module / ModuleCluster records as follows:
	//
	// 1. The processor was the last to support the cluster in the module
	//		=> the cluster record is removed from the module
	// 2. The processor was the last to support the module
	//		=> the module record is removed from the table
	// 3. The processor was not the last to support the module or cluster
	//		=> the processor is removed from the cluster records processor list
	//
	for moduleIdentifier, modules := range table.modules {

		for clusterIdentifier, cluster := range modules.clusters {

			jdx := 0
			var processor *Processor = nil
			for jdx, processor = range cluster.processors {
				// compare the pointers
				if processor == instance {
					break
				}
			}

			cluster.processors = append(cluster.processors[:jdx], cluster.processors[jdx+1:]...)

			if len(cluster.processors) == 0 {
				delete(modules.clusters, clusterIdentifier)
			}
		}

		if len(modules.clusters) == 0 {
			delete(table.modules, moduleIdentifier)
		}
	}

	return nil
}

// GetModule
// n/a
func (table *Table) GetModule(name string) (instance *Module, found bool) {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	instance, found = table.modules[name]
	return instance, found
}

// AddModule
// inform the core that the processor now supports provisioning calls
// for a module and all its listed clusters
func (table *Table) AddModule(processorName string, config *interfaces.ModuleConfig) error {

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

	/* addCluster the module name to the provisioner for reference */
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
		if clusterInstance, found := moduleInstance.GetCluster(export.Cluster); found {
			clusterInstance.Add(processorInstance)
			continue
		}

		/* if the cluster doesn't exist this is the first time we will have the record */
		moduleInstance.addCluster(export.Cluster)
		clusterInstance, _ := moduleInstance.GetCluster(export.Cluster)

		/* associate the processor as one of the executors for this cluster */
		clusterInstance.Add(processorInstance)

		/* if this is the first time creating this cluster, we should follow the default
		   mount request outlined by the module config
		*/
		if export.StaticMount {
			clusterInstance.Mount()
		}

		clusterInstance.data.Mode = export.Config.Mode
	}

	// TODO : allow the user to specify whether they want modules to be mounted by default
	// for now modules will be mounted by default to make docker deployments easier
	moduleInstance.Mount()

	table.modules[config.Name] = moduleInstance

	return nil
}

// RemoveModule
// remove a module from a processor
func (table *Table) RemoveModule(processor, name string) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	var instance *Processor
	for _, instance = range table.processors {

		if instance.ToString() == processor {
			break
		}
	}

	if instance == nil {
		return errors.New("processor does not exist")
	}

	module, found := table.modules[name]

	if !found {
		return errors.New("module does not exist")
	}

	for clusterIdentifier, cluster := range module.clusters {

		for idx, processor := range cluster.processors {

			if processor == instance {
				cluster.processors = append(cluster.processors[:idx], cluster.processors[idx+1:]...)
				break
			}
		}

		if len(cluster.processors) == 0 {
			delete(module.clusters, clusterIdentifier)
		}
	}

	if len(module.clusters) == 0 {
		delete(table.modules, name)
	}

	for idx, module := range instance.Modules {

		if module == name {
			instance.Modules = append(instance.Modules[:idx], instance.Modules[idx+1:]...)
		}
	}

	return nil
}

// RegisteredModules
// Fetch a copy of all modules stored on the core.
func (table *Table) RegisteredModules() []ModuleData {

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

// Print
// visual representation of the state of the table
func (table *Table) Print() {

	for identifier, module := range table.modules {
		fmt.Printf("├─ %s (mounted: %t) \n", identifier, module.IsMounted())

		for identifier, cluster := range module.clusters {

			fmt.Printf("|  ├─%s (mounted: %t)\n", identifier, cluster.IsMounted())

			for _, processor := range cluster.processors {
				fmt.Printf("|  |  ├─%s\n", processor.ToString())
			}
		}
	}
}
