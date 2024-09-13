package processor

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/api"
	"github.com/GabeCordo/cluster-tools/internal/processor"
)

func (t *Thread) processorGet() []*processor.Processor {

	return t.processorTable.GetProcessors()
}

func (t *Thread) processorAdd(config *processor.Config) error {

	err := t.processorTable.AddProcessor(config)

	if err == nil {
		t.Logger.Printf("[%s:%d -> cluster-tools] connected a new processor\n",
			config.Host, config.Port)
	} else {
		t.Logger.Printf("[%s:%d -> cluster-tools] received a processor connection but there was a failure\n%s\n",
			config.Host, config.Port, err.Error())
	}
	return err
}

func (t *Thread) processorRemove(config *processor.Config) error {

	err := t.processorTable.RemoveProcessor(config)

	if err == nil {
		t.Logger.Printf("[%s:%d -> cluster-tools] disconnected a processor\n",
			config.Host, config.Port)
		t.processorTable.Print()
	} else {
		t.Logger.Printf("[%s:%d -> cluster-tools] received a processor disconnected but there was a failure\n%s\n",
			config.Host, config.Port, err.Error())
	}
	return err
}

func (t *Thread) processorPing() {

	table := t.processorTable
	processors := table.GetProcessors()

	// iterate over each processor and probe whether they are still
	// reachable, if not, the processor state should be updated
	for _, p := range processors {

		// the processor probe failed if err is not nil
		if err := api.Probe(p); err != nil {

			var suffix string
			if p.Retries > 0 {
				suffix = fmt.Sprintf(" (retry %d)", p.Retries)
			}

			t.Logger.Printf("[cluster-tools -> %s:%d] unable to probe processor %s\n", p.Host, p.Port, suffix)

			if (p.Retries + 1) >= t.config.MaxRetry {

				t.Logger.Printf("max probe retries hit, removing processor %s:%d\n", p.Host, p.Port)

				if err = table.RemoveProcessor(&processor.Config{
					Host: p.Host,
					Port: p.Port,
				}); err != nil {
					t.Logger.Println("failed to remove processor")
				}
			} else {
				p.Retries++
			}
		} else {
			p.Retries = 0
		}
	}
}
