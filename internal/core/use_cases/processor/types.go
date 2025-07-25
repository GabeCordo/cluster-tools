package processor

import (
	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/toolchain/logging"
)

type UseCases struct {
	ProcessorTable *processor.Table
	Logger         *logging.Logger
}
