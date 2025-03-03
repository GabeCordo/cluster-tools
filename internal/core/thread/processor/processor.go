package processor

import (
	"github.com/Sentmint/pops/internal/core/processor"
)

func (t *Thread) processorGet() []*processor.Processor {

	return t.processorTable.GetProcessors()
}

func (t *Thread) processorAdd(config *processor.Config) error {

	_, err := t.processorTable.AddProcessor(config)
	if err == nil {
		t.Logger.Printf("[%s -> pops-core] connected a new processor\n",
			config.RemoteAddr)
	} else {
		t.Logger.Printf("[%s -> pops-core] received a processor connection but there was a failure\n%s\n",
			config.RemoteAddr, err.Error())
	}

	return err
}

func (t *Thread) processorRemove(config *processor.Config) error {

	err := t.processorTable.RemoveProcessor(config)

	if err == nil {
		t.Logger.Printf("[%s -> pops-core] disconnected a processor\n",
			config.RemoteAddr)
		t.processorTable.Print()
	} else {
		t.Logger.Printf("[%s -> pops-core] received a processor disconnected but there was a failure\n%s\n",
			config.RemoteAddr, err.Error())
	}
	return err
}
