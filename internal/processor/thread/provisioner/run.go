package provisioner

import (
	"errors"
	"sync"
	"time"

	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/processor/component/provision"
	"github.com/GabeCordo/Flock/internal/processor/component/provision/pipeline"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	"github.com/GabeCordo/toolchain/logging"
)

func (t *Thread) getSupervisor() []*provision.Run {

	return nil
}

func (t *Thread) provisionRun(request *thread.ProvisionerRequest) error {

	// Note: configs are now sent from the core, we don't need to worry about looking for, verifying, or
	//		 reverting to a default cluster.pipeline if one is not provided
	if request == nil || request.Pipeline == nil {
		return errors.New("passed value is nil")
	}

	// if we have exceeded the number of supervisors we want to statistic on the processor,
	// we can add it to a queue that will be statistic at another time.
	//
	// note: we can tell the core pre-maturely that the runner was provisioned so
	// 		 that the caller is told that the request successfully reached this server.
	if t.NumOfActiveSupervisors() >= MaxNumOfSupervisors {
		t.requestBacklog = append(t.requestBacklog, request)
		return nil
	} else {
		t.IncrementActiveSupervisors()
	}

	// TODO : cleanup
	t.logger.Printf("%s[%s]%s Provisioning pipeline in module %s\n", logging.Green, request.Pipeline.Identifier, logging.Reset, "foo")

	supervisorInstance, err := t.provisioner.CreateSupervisor(
		request.Namespace, request.Supervisor, request.Metadata, *t.Config.Core, request.Pipeline)

	if err != nil {
		t.logger.Printf("%s[%s]%s Failed to create runner %s\n", logging.Red, request.Pipeline.Identifier, logging.Reset, err)
		return err
	}

	t.logger.Printf("%s[%s]%s Pipeline Active (run: %d)\n", logging.Green, request.Pipeline.Identifier, logging.Reset, supervisorInstance.Id)
	go func(supervisorInstance *pipeline.Instance) {

		m := sync.Mutex{} // used for sending updates to the gateway

		// premise:
		// the statistics associated with the running pipeline instance will be updated on-demand
		// as data flows through the channels between functions.
		//
		// idea:
		// every 1s send an update of the statistics to the flock gateway so the operator
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
				if !supervisorInstance.IsAlive() {
					// the supervisor is expected to leave the 'alive' state at an
					// undefined point in its run, this is the exit-case for the background loop
					break
				}

				r := new(run.Run)
				if r == nil {
					panic("failed to allocate memory for run.Run")
				}

				r.Id = supervisorInstance.Id
				r.Status = run.Active
				r.Statistics = supervisorInstance.Pipeline.Stats

				mandatory := thread.SocketMandatory{Pipe: t.channels.C0, NoncePool: t.noncePool}
				thread.AsyncRunUpdate(mandatory, r)

				m.Unlock()

				time.Sleep(1 * time.Second) // wait before the next update
			}
		}()

		t.runWg.Add(1)

		// block until the runner completes
		supervisorInstance.Start()

		m.Lock()

		status := string(supervisorInstance.State)

		r := new(run.Run)
		if r == nil {
			panic("failed to allocate memory for run.Run")
		}

		r.Id = supervisorInstance.Id
		r.Status = run.FromString(status)
		r.Statistics = supervisorInstance.Pipeline.Stats

		mandatory := thread.SocketMandatory{Pipe: t.channels.C0, NoncePool: t.noncePool}
		thread.AsyncRunUpdate(mandatory, r)

		m.Unlock()

		// provide the console with output indicating that the cluster has completed
		// we already provide output when a cluster is provisioned, so it completes the state
		if *t.Config.Debug {
			duration := time.Now().Sub(supervisorInstance.StartTime)
			t.logger.Printf("%s[%s]%s Pipeline complete, took %dhr %dm %ds %dms %dus\n",
				logging.Green,
				supervisorInstance.Pipeline.Identifier,
				logging.Reset,
				int(duration.Hours()),
				int(duration.Minutes()),
				int(duration.Seconds()),
				int(duration.Milliseconds()),
				int(duration.Microseconds()),
			)
		}

		// once a runner has sent its data to the core there is no reason to keep the data
		// stored in memory without risking it staying unused till the program is terminated
		deleted, _ := t.provisioner.DeleteSupervisor(supervisorInstance.Id)
		if deleted {
			t.logger.Printf("deleted runner %d\n", supervisorInstance.Id)
		} else {
			t.logger.Warnf("failed to delete runner %d\n", supervisorInstance.Id)
		}

		// let the modules t decrement the semaphore otherwise we will be stuck in deadlock waiting for
		// the provisioned cluster to complete before allowing the etl-t to shut down
		//if !clusterWrapper.IsStream() {
		t.DecrementActiveSupervisors()
		t.runWg.Done()
	}(supervisorInstance)

	return nil
}

func (t *Thread) stopRun(request *thread.ProvisionerRequest) error {

	instance, found := t.provisioner.GetSupervisor(request.Supervisor)
	if !found {
		return errors.New("no supervisor with that id exists")
	}

	instance.Teardown()
	return nil
}
