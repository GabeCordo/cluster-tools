package database

import "github.com/GabeCordo/Flock/internal/core/database"

type UseCases struct {
	PipelineDatabase  database.Database
	StatisticDatabase database.Database
	JobDatabase       database.Database
}
