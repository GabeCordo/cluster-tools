package provisioner

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/processor/api"
	"github.com/GabeCordo/cluster-tools/internal/processor/provision"
	"github.com/GabeCordo/cluster-tools/internal/processor/provision/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/toolchain/logging"
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

	thread.logger.Printf("%s[%s]%s Pipeline Running (run: %d)\n", logging.Green, request.Pipeline.Identifier, logging.Reset, supervisorInstance.Id)
	go func(supervisorInstance *pipeline.Instance) {

		// block until the runner completes
		response := supervisorInstance.Start()
		// TODO : should we send the response instead?

		if !*thread.Config.Standalone {
			api.UpdateRun(*thread.Config.Core, supervisorInstance.Id, provision.RunStatus(supervisorInstance.State), response.Stats)
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
