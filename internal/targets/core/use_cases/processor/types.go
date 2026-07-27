package processor

import (
	"github.com/GabeCordo/FunctionScheduler/internal/shared/logging"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/component/processor"
)

type UseCases struct {
	ProcessorTable *processor.Table
	Logger         logging.Logger
}
