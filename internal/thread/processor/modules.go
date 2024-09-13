package processor

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/database/config"
	"github.com/GabeCordo/cluster-tools/internal/processor"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"math/rand"
)

func (t *Thread) getModules() []processor.ModuleData {

	return t.processorTable.RegisteredModules()
}

func (t *Thread) addModule(processorName string, cfg *processor.ModuleConfig) error {

	if !cfg.Verify() {
		return errors.New("module config is not valid")
	}

	if err := t.processorTable.AddModule(processorName, cfg); err != nil {
		return err
	}

	// the module will send a default config for every cluster it registers within it
	// this config should be used as the de-facto config unless another is specified by the operator
	// -> send the config for storage in the database t
	for _, export := range cfg.Exports {
		if export.Config.Mode == config.Stream {
			t.C13 <- thread.Request{
				Action: thread.CreateAction,
				Type:   thread.SupervisorRecord,
				Identifiers: thread.RequestIdentifiers{
					Processor: processorName,
					Module:    cfg.Name,
					Cluster:   export.Cluster,
					Config:    export.Cluster,
				},
				Caller: thread.System,
				Data:   make(map[string]string),
				Nonce:  rand.Uint32(),
			}
		}

		mandatory := thread.Mandatory{t.C11, t.DatabaseResponseTable, t.config.Timeout}
		err := thread.StoreConfigInDatabase(mandatory, cfg.Name, export.ToClusterConfig())
		if err == nil {
			t.Logger.Printf("stored new default config for cluster %s in database\n", export.Cluster)
		} else {
			// the config could have already been stored in a previous module register
			// note: configs are not deleted when the processor is disconnected at the moment
			//		-> the idea is we can re-use them s.t. performance can be improved
			fmt.Println(err)
			t.Logger.Printf("failed to database default config for cluster %s in database\n", export.Cluster)
		}
	}

	// let the operator have an understanding of the cluster-tools's state
	// ->	when a processor is added it may change what modules/configs/processors are available to use
	//		and whether they are mounted in the cluster-tools currently
	t.processorTable.Print()

	return nil
}

func (t *Thread) deleteModule(processorName, moduleName string) error {

	return t.processorTable.RemoveModule(processorName, moduleName)
}

func (t *Thread) mountModule(name string) error {

	instance, found := t.processorTable.GetModule(name)
	if !found {
		return processor.ModuleDoesNotExist
	}

	instance.Mount()
	t.Logger.Printf("the module %s was MOUNTED\n", name)
	t.processorTable.Print()

	return nil
}

func (t *Thread) unmountModule(name string) error {

	instance, found := t.processorTable.GetModule(name)
	if !found {
		return processor.ModuleDoesNotExist
	}

	instance.Unmount()
	t.Logger.Printf("the module %s was UNMOUNTED\n", name)
	t.processorTable.Print()

	return nil
}
