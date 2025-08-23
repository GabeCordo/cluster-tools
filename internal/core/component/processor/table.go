package processor

import (
	"errors"
	"fmt"
	"github.com/FortifiedCode/plover"
	"sync"
)

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

func (table *Table) GetProcessors() []*Processor {

	table.mutex.RLock()
	defer table.mutex.RUnlock()

	return table.processors
}

func (table *Table) AddProcessor(cfg *Config) (*Processor, error) {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	for _, processor := range table.processors {
		if processor.Id == cfg.Identifier {
			return nil, AlreadyExists
		}
	}

	processor := newProcessor(cfg.Identifier, cfg.RemoteAddr)
	table.processors = append(table.processors, processor)
	table.NumOfProcessors++

	return processor, nil
}

// RemoveProcessor
// this is a REALLY expensive operation that might need to be optimized in the future.
func (table *Table) RemoveProcessor(cfg *Config) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	idx := 0
	var instance *Processor = nil
	for idx, instance = range table.processors {
		if instance.Id == cfg.Identifier {
			break
		}
	}

	if instance == nil {
		return errors.New("processor does not exist")
	}

	table.processors = append(table.processors[:idx], table.processors[idx+1:]...)

	// Removing a processor modifies the Namespace / ModuleCluster records as follows:
	//
	// 1. The processor was the last to support the cluster in the module
	//		=> the cluster record is removed from the module
	// 2. The processor was the last to support the module
	//		=> the module record is removed from the table
	// 3. The processor was not the last to support the module or cluster
	//		=> the processor is removed from the cluster records processor list
	//
	for moduleIdentifier, modules := range table.modules {

		for clusterIdentifier, cluster := range modules.functions {

			// TODO : this is a hack fix

			jdx := 0
			var processor *Processor = nil
			for jdx, processor = range cluster.Processors {
				// compare the pointers
				if processor == instance {
					break
				}
			}

			cluster.Processors = append(cluster.Processors[:jdx], cluster.Processors[jdx+1:]...)

			if len(cluster.Processors) == 0 {
				delete(modules.functions, clusterIdentifier)
			}
		}

		if len(modules.functions) == 0 {
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
// inform the flock that the processor now supports provisioning calls
// for a module and all its listed functions
func (table *Table) AddModule(processorId uint64, config *plover.ModuleIR) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	var processorInstance *Processor
	for _, instance := range table.processors {
		if instance.Id == processorId {
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
		if module == config.Identifier {
			return ModuleAlreadyRegistered
		}
	}

	/* addFunction the module name to the provisioner for reference */
	processorInstance.Modules = append(processorInstance.Modules, config.Identifier)

	var moduleInstance *Module

	/* if the module already exists we should try to re-use the existing module allocation */
	if instance, found := table.modules[config.Identifier]; found {

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
		moduleInstance = newModule(config.Identifier, config.Version, config.Contact)
	}

	for _, export := range config.Functions {

		/* does the cluster association already exist in the module? */
		/* Note: this can be the case if the module already existed */
		if clusterInstance, found := moduleInstance.GetFunction(export.Identifier); found {
			clusterInstance.Add(processorInstance)
			continue
		}

		/* if the cluster doesn't exist this is the first time we will have the record */
		moduleInstance.addFunction(&export)
		clusterInstance, _ := moduleInstance.GetFunction(export.Identifier)

		/* associate the processor as one of the executors for this cluster */
		clusterInstance.Add(processorInstance)

		/* if this is the first time creating this cluster, we should follow the default
		   mount request outlined by the module pipeline
		*/
		if export.Metadata.StaticMount {
			clusterInstance.Mount()
		}

		// TODO : remove
		//clusterInstance.SetMode(export.Pipeline.Mode)
	}

	// TODO : allow the user to specify whether they want modules to be mounted by default
	// for now modules will be mounted by default to make docker deployments easier
	moduleInstance.Mount()

	table.modules[config.Identifier] = moduleInstance

	return nil
}

// RemoveModule
// remove a module from a processor
func (table *Table) RemoveModule(processor uint64, name string) error {

	table.mutex.Lock()
	defer table.mutex.Unlock()

	var instance *Processor
	for _, instance = range table.processors {
		if instance.Id == processor {
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

	for clusterIdentifier, cluster := range module.functions {

		for idx, processor := range cluster.Processors {

			if processor == instance {
				cluster.Processors = append(cluster.Processors[:idx], cluster.Processors[idx+1:]...)
				break
			}
		}

		if len(cluster.Processors) == 0 {
			delete(module.functions, clusterIdentifier)
		}
	}

	if len(module.functions) == 0 {
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
// Fetch a copy of all modules stored on the flock.
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

		for identifier, cluster := range module.functions {

			fmt.Printf("|  ├─%s (mounted: %t)\n", identifier, cluster.IsMounted())

			for _, processor := range cluster.Processors {
				fmt.Printf("|  |  ├─%s\n", processor.ToString())
			}
		}
	}
}
