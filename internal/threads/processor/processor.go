package processor

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/api"
	"github.com/GabeCordo/cluster-tools/internal/core/components/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
)

func (thread *Thread) processorGet() []*processor.Processor {

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

func (thread *Thread) processorPing() {

	table := GetTableInstance()
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

			thread.Logger.Printf("[core -> %s:%d] unable to probe processor %s\n", p.Host, p.Port, suffix)

			if (p.Retries + 1) >= thread.config.MaxRetry {

				thread.Logger.Printf("max probe retries hit, removing processor %s:%d\n", p.Host, p.Port)

				if err = table.RemoveProcessor(&interfaces.ProcessorConfig{
					Host: p.Host,
					Port: p.Port,
				}); err != nil {
					thread.Logger.Println("failed to remove processor")
				}
			} else {
				p.Retries++
			}
		} else {
			p.Retries = 0
		}
	}
}
