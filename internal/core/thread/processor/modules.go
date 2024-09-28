package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
)

func (t *Thread) getModules() []processor.ModuleData {

	return t.processorTable.RegisteredModules()
}

func (t *Thread) addModule(processorName string, cfg *processor.ModuleConfig) error {

	if !cfg.Verify() {
		return errors.New("module pipeline is not valid")
	}

	if err := t.processorTable.AddModule(processorName, cfg); err != nil {
		return err
	}

	// the module will send a default pipeline for every cluster it registers within it
	// this pipeline should be used as the de-facto pipeline unless another is specified by the operator
	// -> send the pipeline for storage in the database t
	//for _, export := range cfg.Exports {
	//	if export.Pipeline.Mode == pipeline.Stream {
	//		t.C13 <- thread.Request{
	//			Action: thread.CreateAction,
	//			Type:   thread.SupervisorRecord,
	//			Identifiers: thread.RequestIdentifiers{
	//				Processor: processorName,
	//				Namespace:    cfg.Name,
	//				Function:   export.Function,
	//				Pipeline:    export.Function,
	//			},
	//			Caller: thread.System,
	//			Data:   make(map[string]string),
	//			Nonce:  rand.Uint32(),
	//		}
	//	}

	// TODO : we are removing configs from processor
	//mandatory := thread.Mandatory{t.C11, t.DatabaseResponseTable, t.pipeline.Timeout}
	//err := thread.StoreConfigInDatabase(mandatory, cfg.Name, export.ToClusterConfig())
	//if err == nil {
	//	t.Logger.Printf("stored new default pipeline for cluster %s in database\n", export.Function)
	//} else {
	//	// the pipeline could have already been stored in a previous module register
	//	// note: configs are not deleted when the processor is disconnected at the moment
	//	//		-> the idea is we can re-use them s.t. performance can be improved
	//	fmt.Println(err)
	//	t.Logger.Printf("failed to database default pipeline for cluster %s in database\n", export.Function)
	//}
	//}

	// let the operator have an understanding of the ctgate's state
	// ->	when a processor is added it may change what modules/configs/processors are available to use
	//		and whether they are mounted in the ctgate currently
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
