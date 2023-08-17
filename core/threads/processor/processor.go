package processor

import (
	"github.com/GabeCordo/etl-light/module"
	processor_i "github.com/GabeCordo/etl-light/processor"
	"github.com/GabeCordo/etl/core/components/processor"
)

func (thread *Thread) processorGet() []processor.Processor {

	return GetTableInstance().GetProcessors()
}

func (thread *Thread) processorAdd(config *processor_i.Config) error {

	return GetTableInstance().AddProcessor(config)
}

func (thread *Thread) processorRemove(config module.Config) {

}
