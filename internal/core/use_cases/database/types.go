package database

import (
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/toolchain/logging"
)

type UseCases struct {
	PipelineDatabase  database.Database
	StatisticDatabase database.Database
	JobDatabase       database.Database
	Logger            *logging.Logger
}
