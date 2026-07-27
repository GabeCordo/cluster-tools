package database

import (
	"github.com/GabeCordo/FunctionScheduler/internal/shared/logging"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/job"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/pipeline"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/statistic"
)

type UseCases struct {
	PipelineDatabase  pipeline.Database
	StatisticDatabase statistic.Database
	JobDatabase       job.Database
	Logger            logging.Logger
}
