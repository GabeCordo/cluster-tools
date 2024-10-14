package processor

import (
	"github.com/Sentmint/cluster-tools/internal/core/processor"
)

func (t *Thread) processorGet() []*processor.Processor {

	return t.processorTable.GetProcessors()
}

func (t *Thread) processorAdd(config *processor.Config) error {

	err := t.processorTable.AddProcessor(config)

	if err == nil {
		t.Logger.Printf("[proc: %d -> ctgate] connected a new processor\n",
			config.Identifier)
	} else {
		t.Logger.Printf("[%s:%d -> ctgate] received a processor connection but there was a failure\n%s\n",
			config.Identifier, err.Error())
	}
	return err
}

func (t *Thread) processorRemove(config *processor.Config) error {

	err := t.processorTable.RemoveProcessor(config)

	if err == nil {
		t.Logger.Printf("[proc: %d -> ctgate] disconnected a processor\n",
			config.Identifier)
		t.processorTable.Print()
	} else {
		t.Logger.Printf("[proc: %d -> ctgate] received a processor disconnected but there was a failure\n%s\n",
			config.Identifier, err.Error())
	}
	return err
}
