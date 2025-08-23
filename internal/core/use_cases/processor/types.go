package processor

import (
	"github.com/FortifiedCode/flock/internal/core/component/processor"
	"github.com/FortifiedCode/flock/internal/shared/logging"
)

type UseCases struct {
	ProcessorTable *processor.Table
	Logger         logging.Logger
}
