package scheduler

import (
	"github.com/FortifiedCode/flock/internal/core/component/scheduler/job"
	jobDatabase "github.com/FortifiedCode/flock/internal/core/database/job"
	"github.com/FortifiedCode/flock/internal/shared/logging"
)

type UseCases struct {
	Scheduler    *job.Scheduler
	JobsDatabase jobDatabase.Database
	Logger       logging.Logger
}
