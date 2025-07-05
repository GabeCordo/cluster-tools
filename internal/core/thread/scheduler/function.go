package scheduler

import (
	"errors"
	"fmt"

	"github.com/GabeCordo/Flock/internal/core/component/processor"
	job2 "github.com/GabeCordo/Flock/internal/core/component/scheduler/job"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/job"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/toolchain/multithreaded"
)

func (t *Thread) watch() {

	job2.Watch(t.Scheduler)
}

func (t *Thread) loop() {

	err := job2.Loop(t.Scheduler, func(jb job.Job) error {

		// will return have a maximum of Timeout, so worst-case takes thread.pipeline.Timeout
		mandatory := thread.Mandatory{
			Pipe:          t.channels.c18,
			ResponseTable: t.processorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		}
		_, err := thread.CreateRun(mandatory, jb.Namespace, jb.Pipeline, jb.Metadata)

		e := ""
		if err != nil {

			e = fmt.Sprintf("but encountered an error, %s", err.Error())
		}

		if (err != nil) && t.config.Debug {
			t.logger.Printf("scheduled cluster is ready: %s (%s,%s) %s\n", jb.Identifier, jb.Namespace, jb.Pipeline, e)
			t.logger.Printf("%d clusters are waiting to be provisioned\n", t.Scheduler.ItemsInQueue())
		}

		// if err is not nil, the Scheduler will stop running, so output to console
		// if debug is enabled so the operator is aware of the runtime change
		if (errors.Is(err, processor.CanNotProvisionStreamCluster) || (errors.Is(err, multithreaded.NoResponseReceived))) && t.config.Debug {
			t.logger.Printf("the Scheduler stopped after encountering %s\n", err.Error())
		}

		// I only care about errors that might indicate a compromised state of the thread, the others
		// like Namespace/Function's not existing really makes no sense to crash the Scheduler as someone
		// likely put in the job for a future module/cluster pair they want to attach to mango
		if errors.Is(err, processor.CanNotProvisionStreamCluster) || errors.Is(err, multithreaded.NoResponseReceived) ||
			errors.Is(err, processor.ModuleDoesNotExist) || errors.Is(err, processor.FunctionDoesNotExist) {
			return err
		} else {
			return nil
		}
	})
	if err != nil {
		fmt.Print(err)
	}
}

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
