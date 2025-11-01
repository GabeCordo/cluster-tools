package scheduler

import (
	"errors"
	"fmt"
	"github.com/FortifiedCode/flock/internal/shared/nonce"
	"github.com/FortifiedCode/flock/internal/targets/core/component/processor"
	job2 "github.com/FortifiedCode/flock/internal/targets/core/component/scheduler/job"
	"github.com/FortifiedCode/flock/internal/targets/core/database"
	"github.com/FortifiedCode/flock/internal/targets/core/database/job"
)

func (uc UseCases) GetJobs(filter database.Filter) (jobs []*job.Job) {

	results := uc.Scheduler.Jobs.Get(filter)

	jobs = make([]*job.Job, len(results))
	for i, result := range results {
		jobs[i] = result
	}

	return jobs
}

func (uc UseCases) CreateJob(newJob *job.Job) error {
	_, err := uc.Scheduler.Jobs.Create(database.Filter{}, newJob)
	return err
}

func (uc UseCases) DeleteJob(filter database.Filter) error {
	return uc.Scheduler.Jobs.Delete(filter)
}

func (uc UseCases) GetQueue() []job.Job {
	return uc.Scheduler.GetQueue()
}

func (uc UseCases) SchedulerWatch() {
	job2.Watch(uc.Scheduler)
}

func (uc UseCases) SchedulerLoop(sendMsg func(namespaceId, pipelineId string, metadata map[string]string) error) {
	err := job2.Loop(uc.Scheduler, func(jb job.Job) error {

		// send an asynchronous message to provision the run defined by the scheduler job
		err := sendMsg(jb.Namespace, jb.Pipeline, jb.Metadata)

		e := ""
		if err != nil {

			e = fmt.Sprintf("but encountered an error, %s", err.Error())
		}

		if err != nil {
			uc.Logger.Printf("scheduled cluster is ready: %s (%s,%s) %s\n", jb.Identifier, jb.Namespace, jb.Pipeline, e)
			uc.Logger.Printf("%d clusters are waiting to be provisioned\n", uc.Scheduler.ItemsInQueue())
		}

		// if err is not nil, the Scheduler will stop running, so output to console
		// if debug is enabled so the operator is aware of the runtime change
		if errors.Is(err, processor.CanNotProvisionStreamCluster) || errors.Is(err, nonce.NoResponseReceived) {
			uc.Logger.Printf("the Scheduler stopped after encountering %s\n", err.Error())
		}

		// I only care about errors that might indicate a compromised state of the thread, the others
		// like Namespace/Function's not existing really makes no sense to crash the Scheduler as someone
		// likely put in the job for a future module/cluster pair they want to attach to mango
		if errors.Is(err, processor.CanNotProvisionStreamCluster) || errors.Is(err, nonce.NoResponseReceived) ||
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

func (uc UseCases) LoadJobsFromDisk(schedulersFolder string) error {

	return uc.Scheduler.Jobs.Load(schedulersFolder)
}

func (uc UseCases) SaveJobsToDisk(schedulersFolder string) {

	err := uc.Scheduler.Jobs.Save(schedulersFolder)
	if err != nil {
		uc.Logger.Warnln(err.Error())
	}
}
