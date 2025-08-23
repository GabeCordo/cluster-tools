package database

import (
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/shared/logging"
)

type UseCases struct {
	PipelineDatabase  database.Database
	StatisticDatabase database.Database
	JobDatabase       database.Database
	Logger            logging.Logger
}
