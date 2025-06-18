package processor

import (
	"github.com/GabeCordo/Flock/internal/core/component/processor"
)

func (t *Thread) syncGetProcessors() []*processor.Processor {

	return t.processorTable.GetProcessors()
}

func (t *Thread) synchAddProcessor(config *processor.Config) error {

	_, err := t.processorTable.AddProcessor(config)
	if err == nil {
		t.Logger.Printf("[%s -> flock] connected a new processor\n",
			config.RemoteAddr)
	} else {
		t.Logger.Printf("[%s -> flock] received a processor connection but there was a failure\n%s\n",
			config.RemoteAddr, err.Error())
	}

	return err
}

func (t *Thread) syncDeleteProcessor(config *processor.Config) error {

	err := t.processorTable.RemoveProcessor(config)

	if err == nil {
		t.Logger.Printf("[%s -> flock] disconnected a processor\n",
			config.RemoteAddr)
		t.processorTable.Print()
	} else {
		t.Logger.Printf("[%s -> flock] received a processor disconnected but there was a failure\n%s\n",
			config.RemoteAddr, err.Error())
	}
	return err
}
