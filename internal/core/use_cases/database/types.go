package database

import (
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/shared/logging"
)

type UseCases struct {
	PipelineDatabase  database.Database
	StatisticDatabase database.Database
	JobDatabase       database.Database
	Logger            logging.Logger
}
