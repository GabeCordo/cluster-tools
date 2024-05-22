package processor

import (
	"github.com/GabeCordo/cluster-tools/internal/core/components/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
)

func (thread *Thread) processorGet() []processor.Processor {

	return GetTableInstance().GetProcessors()
}

func (thread *Thread) processorAdd(config *interfaces.ProcessorConfig) error {

	err := GetTableInstance().AddProcessor(config)

	if err == nil {
		thread.Logger.Printf("[%s:%d -> core] connected a new processor\n",
			config.Host, config.Port)
	} else {
		thread.Logger.Printf("[%s:%d -> core] received a processor connection but there was a failure\n%s\n",
			config.Host, config.Port, err.Error())
	}
	return err
}

func (thread *Thread) processorRemove(config *interfaces.ProcessorConfig) error {

	err := GetTableInstance().RemoveProcessor(config)

	if err == nil {
		thread.Logger.Printf("[%s:%d -> core] disconnected a processor\n",
			config.Host, config.Port)
		GetTableInstance().Print()
	} else {
		thread.Logger.Printf("[%s:%d -> core] received a processor disconnected but there was a failure\n%s\n",
			config.Host, config.Port, err.Error())
	}
	return err
}
