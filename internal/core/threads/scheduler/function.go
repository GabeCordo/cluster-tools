package scheduler

import (
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
)

func (thread *Thread) get(filter *interfaces.Filter) []interfaces.Job {

	jobs, _ := thread.Scheduler.Jobs.GetBy(filter)
	return jobs
}

func (thread *Thread) create(job *interfaces.Job) error {

	return thread.Scheduler.Jobs.Create(job)
}

func (thread *Thread) delete(filter *interfaces.Filter) error {

	return thread.Scheduler.Jobs.Delete(filter)
}

func (thread *Thread) queue() []interfaces.Job {

	return thread.Scheduler.GetQueue()
}
