package scheduler

import (
	"github.com/GabeCordo/Flock/internal/core/component/scheduler/job"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/toolchain/logging"
)

type UseCases struct {
	Scheduler    *job.Scheduler
	JobsDatabase database.Database
	Logger       *logging.Logger
}
