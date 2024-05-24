package scheduler

import (
	"github.com/GabeCordo/cluster-tools/core/components/scheduler"
)

func (thread *Thread) get(filter *scheduler.Filter) []scheduler.Job {

	return thread.Scheduler.GetBy(filter)
}

func (thread *Thread) create(job *scheduler.Job) error {

	return thread.Scheduler.Create(job)
}

func (thread *Thread) delete(filter *scheduler.Filter) error {

	return thread.Scheduler.Delete(filter)
}

func (thread *Thread) queue() []scheduler.Job {

	return thread.Scheduler.GetQueue()
}
