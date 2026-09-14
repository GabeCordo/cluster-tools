package database

import (
	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/pipeline"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic"
)

type UseCases struct {
	PipelineDatabase  pipeline.Database
	StatisticDatabase statistic.Database
	JobDatabase       job.Database
	Logger            logging.Logger
}
