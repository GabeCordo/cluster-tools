package processor

import (
	"github.com/FortifiedCode/flock/internal/shared/logging"
	"github.com/FortifiedCode/flock/internal/targets/core/component/processor"
)

type UseCases struct {
	ProcessorTable *processor.Table
	Logger         logging.Logger
}
