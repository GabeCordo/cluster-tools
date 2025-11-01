package database

import (
	"github.com/FortifiedCode/flock/internal/shared/logging"
	"github.com/FortifiedCode/flock/internal/targets/core/database/job"
	"github.com/FortifiedCode/flock/internal/targets/core/database/pipeline"
	"github.com/FortifiedCode/flock/internal/targets/core/database/statistic"
)

type UseCases struct {
	PipelineDatabase  pipeline.Database
	StatisticDatabase statistic.Database
	JobDatabase       job.Database
	Logger            logging.Logger
}
