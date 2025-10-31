package database

import (
	"github.com/FortifiedCode/flock/internal/core/database/job"
	"github.com/FortifiedCode/flock/internal/core/database/pipeline"
	"github.com/FortifiedCode/flock/internal/core/database/statistic"
	"github.com/FortifiedCode/flock/internal/shared/logging"
)

type UseCases struct {
	PipelineDatabase  pipeline.Database
	StatisticDatabase statistic.Database
	JobDatabase       job.Database
	Logger            logging.Logger
}
