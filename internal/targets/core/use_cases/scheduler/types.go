package scheduler

import (
	"github.com/GabeCordo/FunctionScheduler/internal/shared/logging"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/component/scheduler/job"
	jobDatabase "github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/job"
)

type UseCases struct {
	Scheduler    *job.Scheduler
	JobsDatabase jobDatabase.Database
	Logger       logging.Logger
}
