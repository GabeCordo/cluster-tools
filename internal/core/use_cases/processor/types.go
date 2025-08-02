package processor

import (
	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/shared/logging"
)

type UseCases struct {
	ProcessorTable *processor.Table
	Logger         logging.Logger
}
