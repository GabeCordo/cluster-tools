package processor

import (
	"github.com/GabeCordo/mango/core/components/processor"
	"github.com/GabeCordo/mango/core/interfaces/module"
	processor_i "github.com/GabeCordo/mango/core/interfaces/processor"
)

func (thread *Thread) processorGet() []processor.Processor {

	return GetTableInstance().GetProcessors()
}

func (thread *Thread) processorAdd(config *processor_i.Config) error {

	return GetTableInstance().AddProcessor(config)
}

func (thread *Thread) processorRemove(config *module.Config) {

}
