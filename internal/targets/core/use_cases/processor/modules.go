package processor

import (
	"errors"
	"fmt"
	processor2 "github.com/FortifiedCode/flock/internal/targets/core/component/processor"
	"github.com/FortifiedCode/plover"
)

func (uc UseCases) GetModules() []processor2.ModuleData {

	return uc.ProcessorTable.RegisteredModules()
}

func (uc UseCases) AddModule(processorId uint64, cfg *plover.ModuleIR) error {

	if ee := plover.VerifyIR(cfg); len(ee) != 0 {
		return errors.New("module pipeline is not valid")
	}

	if err := uc.ProcessorTable.AddModule(processorId, cfg); err != nil {
		return err
	}

	// the module will send a default pipeline for every cluster it registers within it
	// this pipeline should be used as the de-facto pipeline unless another is specified by the operator
	// -> send the pipeline for storage in the database t
	//for _, export := range cfg.Exports {
	//	if export.Data.Mode == pipeline.Stream {
	//		t.c13 <- thread.Request{
	//			Action: thread.CreateAction,
	//			Type:   thread.SupervisorRecord,
	//			Identifiers: thread.RequestIdentifiers{
	//				Processor: processorName,
	//				Namespace:    cfg.Name,
	//				Function:   export.Function,
	//				Data:    export.Function,
	//			},
	//			Caller: thread.System,
	//			Data:   make(map[string]string),
	//			Nonce:  rand.Uint32(),
	//		}
	//	}

	// TODO : we are removing configs from processor
	//mandatory := thread.Mandatory{t.c11, t.DatabaseResponseTable, t.pipeline.Timeout}
	//err := thread.StoreConfigInDatabase(mandatory, cfg.Name, export.ToClusterConfig())
	//if err == nil {
	//	t.logger.Printf("stored new default pipeline for cluster %s in database\n", export.Function)
	//} else {
	//	// the pipeline could have already been stored in a previous module register
	//	// note: configs are not deleted when the processor is disconnected at the moment
	//	//		-> the idea is we can re-use them s.t. performance can be improved
	//	fmt.Println(err)
	//	t.logger.Printf("failed to database default pipeline for cluster %s in database\n", export.Function)
	//}
	//}

	// let the operator have an understanding of the flock's state
	// ->	when a processor is added it may change what modules/configs/processors are available to use
	//		and whether they are mounted in the flock currently
	fmt.Println("UPDATED ==================>")
	uc.ProcessorTable.Print()

	return nil
}

func (uc UseCases) DeleteModule(processorName uint64, moduleName string) error {

	return uc.ProcessorTable.RemoveModule(processorName, moduleName)
}

func (uc UseCases) MountModule(name string) error {

	instance, found := uc.ProcessorTable.GetModule(name)
	if !found {
		return processor2.ModuleDoesNotExist
	}

	instance.Mount()
	uc.Logger.Printf("the module %s was MOUNTED\n", name)
	uc.ProcessorTable.Print()

	return nil
}

func (uc UseCases) UnMountModule(name string) error {

	instance, found := uc.ProcessorTable.GetModule(name)
	if !found {
		return processor2.ModuleDoesNotExist
	}

	instance.Unmount()
	uc.Logger.Printf("the module %s was UNMOUNTED\n", name)
	uc.ProcessorTable.Print()

	return nil
}
