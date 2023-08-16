package processor

import "github.com/GabeCordo/etl/core/components/processor"

var processorTableInstance *processor.Table

func GetTableInstance() *processor.Table {

	if processorTableInstance == nil {
		processorTableInstance = processor.NewTable()
	}
	return processorTableInstance
}
