package scheduler

import (
	"github.com/FortifiedCode/flock/internal/core/component/scheduler/job"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/shared/logging"
)

type UseCases struct {
	Scheduler    *job.Scheduler
	JobsDatabase database.Database
	Logger       logging.Logger
}
