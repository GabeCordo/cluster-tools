package processor

import (
	component "github.com/FortifiedCode/flock/internal/targets/core/component/processor"
)

func (uc UseCases) GetProcessors() []*component.Processor {

	return uc.ProcessorTable.GetProcessors()
}

func (uc UseCases) AddProcessor(config *component.Config) error {

	_, err := uc.ProcessorTable.AddProcessor(config)
	if err == nil {
		uc.Logger.Printf("[%s -> flock] connected a new processor\n",
			config.RemoteAddr)
	} else {
		uc.Logger.Printf("[%s -> flock] received a processor connection but there was a failure\n%s\n",
			config.RemoteAddr, err.Error())
	}

	return err
}

func (uc UseCases) DeleteProcessor(config *component.Config) error {

	err := uc.ProcessorTable.RemoveProcessor(config)

	if err == nil {
		uc.Logger.Printf("[%s -> flock] disconnected a processor\n",
			config.RemoteAddr)
		uc.ProcessorTable.Print()
	} else {
		uc.Logger.Printf("[%s -> flock] received a processor disconnected but there was a failure\n%s\n",
			config.RemoteAddr, err.Error())
	}
	return err
}
