package provisioner

import (
	"errors"
	"github.com/GabeCordo/ScalingFunctions"
	"sync"
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/terminal"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/run"
)

func (useCases *UseCases) GetStatistics() (statistics []*ScalingFunctions.Statistics) {

	runs := useCases.Provisioner.GetRuns()
	statistics = make([]*ScalingFunctions.Statistics, len(runs))

	for _, r := range runs {
		statistics = append(statistics, r.GetStatistics())
	}

	return statistics
}

func (useCases *UseCases) CreateRun(request *ProvisionRequest, updateRunEvent func(r *run.Run)) error {

	// if we have exceeded the number of supervisors we want to statistic on the processor,
	// we can add it to a queue that will be statistic at another time.
	//
	// note: we can tell the core pre-maturely that the runner was provisioned so
	// 		 that the caller is told that the request successfully reached this server.
	if useCases.activeRuns.Load() >= MaxNumOfSupervisors {
		err := useCases.Backlog.Add(request)
		if err != nil {
			return err
		}
	} else {
		useCases.activeRuns.Add(1)
	}

	// TODO : cleanup
	useCases.Logger.Printf("%s[%s]%s Provisioning pipeline in module %s\n", terminal.Green, request.Pipeline.Identifier, terminal.Reset, "foo")

	rInstance, err := useCases.Provisioner.CreateRun(
		request.Namespace, request.Supervisor, request.Metadata, request.Core, request.Pipeline)

	if err != nil {
		useCases.Logger.Printf("%s[%s]%s Failed to create runner %s\n", terminal.Red, request.Pipeline.Identifier, terminal.Reset, err)
		return err
	}

	useCases.Logger.Printf("%s[%s]%s Data Active (run: %d)\n", terminal.Green, request.Pipeline.Identifier, terminal.Reset, rInstance.Id)
	go func(rInstance ScalingFunctions.Interactable) {

		m := sync.Mutex{} // used for sending updates to the gateway

		// premise:
		// the statistics associated with the running pipeline instance will be updated on-demand
		// as data flows through the channels between functions.
		//
		// idea:
		// every 1s send an update of the statistics to the DistributedFunctions gateway so the operator
		// or developer can track the progress of the pipeline instance in real-time
		//
		// important note:
		// we need a mutex because there are 2 instances inside the code where the gateway can be updated
		// with the status of the pipeline instance:
		// 	1) here
		//  2) bellow; when the pipeline instance has stopped
		// the mutex ensures there is (no) data race between the two and the statistic updates in the block
		// bellow are only sent in the pipeline instance has not come to a stop.
		go func() {
			for {
				m.Lock()
				if !rInstance.IsRunning() {
					// the supervisor is expected to leave the 'alive' state at an
					// undefined point in its run, this is the exit-case for the background loop
					break
				}

				r := new(run.Run)
				if r == nil {
					panic("failed to allocate memory for run.Run")
				}

				r.Id = rInstance.Id
				r.Status = run.Active
				r.Statistics = rInstance.GetStatistics()

				updateRunEvent(r)

				m.Unlock()

				time.Sleep(1 * time.Second) // wait before the next update
			}
		}()

		useCases.runWg.Add(1)

		// block until the runner completes
		err = rInstance.Run()
		if err != nil {
			useCases.Logger.Warnln(err.Error())
		}

		m.Lock()

		status := string(rInstance.GetStatus())

		rStartTime := time.Now()

		r := new(run.Run)
		if r == nil {
			panic("failed to allocate memory for run.Run")
		}

		r.Id = rInstance.Id
		r.Status = run.FromString(status)
		r.Statistics = rInstance.GetStatistics()

		updateRunEvent(r)

		m.Unlock()

		// provide the console with output indicating that the cluster has completed
		// we already provide output when a cluster is provisioned, so it completes the state
		duration := time.Now().Sub(rStartTime)
		useCases.Logger.Printf("%s[%s]%s Data complete, took %dhr %dm %ds %dms %dus\n",
			terminal.Green,
			rInstance.Pipeline,
			terminal.Reset,
			int(duration.Hours()),
			int(duration.Minutes()),
			int(duration.Seconds()),
			int(duration.Milliseconds()),
			int(duration.Microseconds()),
		)

		// once a runner has sent its data to the core there is no reason to keep the data
		// stored in memory without risking it staying unused till the program is terminated
		deleted, _ := useCases.Provisioner.DeleteRun(rInstance.Id)
		if deleted {
			useCases.Logger.Printf("deleted runner %d\n", rInstance.Id)
		} else {
			useCases.Logger.Warnf("failed to delete runner %d\n", rInstance.Id)
		}

		// let the modules t decrement the semaphore otherwise we will be stuck in deadlock waiting for
		// the provisioned cluster to complete before allowing the etl-t to shut down
		//if !clusterWrapper.IsStream() {
		useCases.activeRuns.Add(-1)
		useCases.runWg.Done()
	}(rInstance)

	return nil
}

func (useCases *UseCases) StopRun(supervisor uint64) error {

	instance, found := useCases.Provisioner.GetRun(supervisor)
	if !found {
		return errors.New("no supervisor with that id exists")
	}

	instance.Stop()
	return nil
}

func (useCases *UseCases) GetModules() []*ScalingFunctions.Module {

	return useCases.Repository.GetModules()
}

func (useCases *UseCases) StopAllRuns() {

	useCases.Provisioner.SuspendRuns()

	// wait for all the running pipelines to complete before tearing down
	//
	// note: wait for the runs (after) async messages have been completed in
	// 		 case there was a pending run request in the async queue.
	useCases.runWg.Wait()
}

func (useCases *UseCases) CheckBacklog(startRun func(request *ProvisionRequest)) {

	if (useCases.activeRuns.Load() < MaxNumOfSupervisors) && (useCases.Backlog.Size() > 0) {
		value, err := useCases.Backlog.Remove()
		if err != nil {
			useCases.Logger.Warnf("failed to remove backlog %s\n", err)
			return
		}
		request := (value).(*ProvisionRequest)
		startRun(request)
	}
}
