package scheduler

import (
	"github.com/Sentmint/cluster-tools/internal/core/database"
	"github.com/Sentmint/cluster-tools/internal/core/database/job"
)

func (t *Thread) get(filter database.Filter) []job.Job {

	results := t.Scheduler.Jobs.Get(filter)

	jobs := make([]job.Job, len(results))
	for i, result := range results {
		jobs[i] = result.(job.Job)
	}

	return jobs
}

func (t *Thread) create(job *job.Job) error {

	_, err := t.Scheduler.Jobs.Create(database.Filter{}, job)
	return err
}

func (t *Thread) delete(filter database.Filter) error {

	return t.Scheduler.Jobs.Delete(filter)
}

func (t *Thread) queue() []job.Job {

	return t.Scheduler.GetQueue()
}
