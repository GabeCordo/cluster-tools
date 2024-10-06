package provisioner

import (
	"errors"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/Sentmint/cluster-tools/internal/processor/api"
	"github.com/Sentmint/cluster-tools/internal/processor/provision"
	"github.com/Sentmint/cluster-tools/internal/processor/provision/pipeline"
	"github.com/Sentmint/cluster-tools/internal/processor/threads"
	"sync"
	"time"
)

func (thread *Thread) getSupervisor() []*provision.Run {

	return nil
}

func (thread *Thread) provisionRun(request *threads.ProvisionerRequest) error {

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
	if thread.NumOfActiveSupervisors() >= MaxNumOfSupervisors {
		thread.requestBacklog = append(thread.requestBacklog, *request)
		thread.requestWg.Done()
		return nil
	} else {
		thread.IncrementActiveSupervisors()
	}

	// TODO : cleanup
	thread.logger.Printf("%s[%s]%s Provisioning pipeline in module %s\n", logging.Green, request.Pipeline.Identifier, logging.Reset, "foo")

	supervisorInstance, err := thread.provisioner.CreateSupervisor(
		request.Namespace, request.Supervisor, request.Metadata, *thread.Config.Core, request.Pipeline)

	if err != nil {
		thread.logger.Printf("%s[%s]%s Failed to create runner %s\n", logging.Red, request.Pipeline.Identifier, logging.Reset, err)
		thread.requestWg.Done()
		return err
	}

	thread.logger.Printf("%s[%s]%s Pipeline Active (run: %d)\n", logging.Green, request.Pipeline.Identifier, logging.Reset, supervisorInstance.Id)
	go func(supervisorInstance *pipeline.Instance) {

		m := sync.Mutex{} // used for sending updates to the gateway

		// premise:
		// the statistics associated with the running pipeline instance will be updated on-demand
		// as data flows through the channels between functions.
		//
		// idea:
		// every 1s send an update of the statistics to the cluster.tools gateway so the operator
		// or developer can track the progress of the pipeline instance in real-time
		//
		// important note:
		// we need a mutex because there are 2 instances inside the code where the gateway can be updated
		// with the status of the pipeline instance:
		// 	1) here
		//  2) bellow; when the pipeline instance has stopped
		// the mutex ensures there is (no) data race between the two and the statistic updates in the block
		// bellow are only sent in the pipeline instance has not come to a stop.
		if !*thread.Config.Standalone {
			go func() {
				for {
					m.Lock()
					if !supervisorInstance.IsAlive() {
						break
					}

					api.UpdateRun(*thread.Config.Core, supervisorInstance.Id, provision.RunStatus(supervisorInstance.State), supervisorInstance.Pipeline.Stats)
					m.Unlock()

					time.Sleep(1 * time.Second) // wait before the next update
				}
			}()
		}

		// block until the runner completes
		response := supervisorInstance.Start()

		if !*thread.Config.Standalone {
			m.Lock()
			api.UpdateRun(*thread.Config.Core, supervisorInstance.Id, provision.RunStatus(supervisorInstance.State), response.Stats)
			m.Unlock()
		}

		// provide the console with output indicating that the cluster has completed
		// we already provide output when a cluster is provisioned, so it completes the state
		if *thread.Config.Debug {
			duration := time.Now().Sub(supervisorInstance.StartTime)
			thread.logger.Printf("%s[%s]%s Pipeline complete, took %dhr %dm %ds %dms %dus\n",
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
		deleted, _ := thread.provisioner.DeleteSupervisor(supervisorInstance.Id)
		if deleted {
			thread.logger.Printf("deleted runner %d\n", supervisorInstance.Id)
		} else {
			thread.logger.Warnf("failed to delete runner %d\n", supervisorInstance.Id)
		}

		// let the modules thread decrement the semaphore otherwise we will be stuck in deadlock waiting for
		// the provisioned cluster to complete before allowing the etl-threads to shut down
		//if !clusterWrapper.IsStream() {
		thread.DecrementActiveSupervisors()
		thread.requestWg.Done()
	}(supervisorInstance)

	return nil
}
