package processor

import (
	"github.com/GabeCordo/etl-light/module"
	"github.com/GabeCordo/etl/core/components/processor"
)

func (thread *Thread) processorGet() []processor.Processor {

	return GetTableInstance().GetProcessors()
}

func (thread *Thread) processorAdd(config module.Config) {

}

func (thread *Thread) processorRemove(config module.Config) {

}

func (thread *Thread) getModules() map[string]bool {

	return GetTableInstance().Registered()
}

func (thread *Thread) getClusters(name string) (map[string]bool, bool) {

	instance, found := GetTableInstance().Get(name)
	if !found {
		return nil, false
	}

	return instance.Registered(), true
}
