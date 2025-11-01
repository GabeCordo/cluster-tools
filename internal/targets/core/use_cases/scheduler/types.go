package scheduler

import (
	"github.com/FortifiedCode/flock/internal/shared/logging"
	"github.com/FortifiedCode/flock/internal/targets/core/component/scheduler/job"
	jobDatabase "github.com/FortifiedCode/flock/internal/targets/core/database/job"
)

type UseCases struct {
	Scheduler    *job.Scheduler
	JobsDatabase jobDatabase.Database
	Logger       logging.Logger
}
