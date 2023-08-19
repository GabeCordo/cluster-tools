package processor

import "github.com/GabeCordo/mango-core/core/components/processor"

var processorTableInstance *processor.Table

func GetTableInstance() *processor.Table {

	if processorTableInstance == nil {
		processorTableInstance = processor.NewTable()
	}
	return processorTableInstance
}
