package processor

import "github.com/GabeCordo/cluster-tools/internal/core/components/processor"

var processorTableInstance *processor.Table

func GetTableInstance() *processor.Table {

	if processorTableInstance == nil {
		processorTableInstance = processor.NewTable()
	}
	return processorTableInstance
}
