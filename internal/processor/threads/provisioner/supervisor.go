package provisioner

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/processor/api"
	"github.com/GabeCordo/cluster-tools/internal/processor/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/processor/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/toolchain/logging"
	"log"
	"time"
)

func (thread *Thread) getSupervisor() []*interfaces.Supervisor {

	return nil
}

func (thread *Thread) provisionSupervisor(request *threads.ProvisionerRequest) error {

	// Note: all mount checks have been moved to the core

	provisionerInstance := GetProvisionerInstance()

	moduleWrapper, found := provisionerInstance.GetModule(request.Module)

	if !found {
		thread.logger.Warnf("%s[%s]%s Module does not exist\n", logging.Green, request.Module, logging.Reset)
		thread.requestWg.Done()
		return errors.New("module not found")
	}

	clusterWrapper, found := moduleWrapper.GetCluster(request.Cluster)

	if !found {
		thread.logger.Warnf("%s[%s]%s Function does not exist\n", logging.Green, request.Cluster, logging.Reset)
		thread.requestWg.Done()
		return errors.New("cluster not found")
	}

	// an operator shall only provision batch etl processes
	// - stream processes are meant to be run by the system when mounted or unmounted
	if (request.Source == threads.User) && clusterWrapper.IsStream() {
		thread.logger.Warnf("%s[%s]%s Could not provision cluster; it's a stream process\n", logging.Green, request.Module, logging.Reset)
		thread.requestWg.Done()
		return errors.New("a stream cluster cannot be provisioned by a user")
	}

	// if we have exceeded the number of supervisors we want to run on the processor,
	// we can add it to a queue that will be run at another time.
	//
	// note: we can tell the core pre-maturely that the supervisor was provisioned so
	// 		 that the caller is told that the request successfully reached this server.
	if thread.NumOfActiveSupervisors() >= MaxNumOfSupervisors {
		if request != nil {
			thread.requestBacklog = append(thread.requestBacklog, *request)
			thread.requestWg.Done()
		} else {
			log.Printf("tried to add request to backlog but found nil pointer request")
		}
		return nil
	} else {
		thread.IncrementActiveSupervisors()
	}

	thread.logger.Printf("%s[%s]%s Provisioning cluster in module %s\n", logging.Green, request.Cluster, logging.Reset, request.Module)

	// Note: configs are now sent from the core, we don't need to worry about looking for, verifying, or
	//		 reverting to a default cluster.pipeline if one is not provided
	if request.Config == nil {
		request.Config = &clusterWrapper.DefaultConfig
	}
	supervisorInstance := clusterWrapper.CreateSupervisor(request.Supervisor, request.Metadata, thread.Config.Core, thread.Config.Standalone, request.Config)

	thread.logger.Printf("%s[%s]%s Supervisor(%d) registered to cluster(%s)\n", logging.Green, request.Cluster, logging.Reset, supervisorInstance.Id, request.Module)

	thread.logger.Printf("%s[%s]%s Function Running\n", logging.Green, request.Cluster, logging.Reset)

	go func(supervisorInstance *supervisor.Supervisor) {

		if !thread.Config.Standalone && clusterWrapper.IsStream() {
			go func() {
				for {
					if !supervisorInstance.IsAlive() {
						break
					} else {
						api.UpdateSupervisor(thread.Config.Core, supervisorInstance.Id, interfaces.SupervisorStatus(supervisorInstance.State), supervisorInstance.Stats)
					}

					time.Sleep(1 * time.Second)
				}
			}()
		}

		// block until the supervisor completes
		response := supervisorInstance.Start()
		// TODO : should we send the response instead?

		// TODO : define host
		if !thread.Config.Standalone {
			api.UpdateSupervisor(thread.Config.Core, supervisorInstance.Id, interfaces.SupervisorStatus(supervisorInstance.State), response.Stats)
		}

		// provide the console with output indicating that the cluster has completed
		// we already provide output when a cluster is provisioned, so it completes the state
		if thread.Config.Debug {
			duration := time.Now().Sub(supervisorInstance.StartTime)
			thread.logger.Printf("%s[%s]%s Function transformations complete, took %dhr %dm %ds %dms %dus\n",
				logging.Green,
				supervisorInstance.Config.Identifier,
				logging.Reset,
				int(duration.Hours()),
				int(duration.Minutes()),
				int(duration.Seconds()),
				int(duration.Milliseconds()),
				int(duration.Microseconds()),
			)
		}

		// once a supervisor has sent its data to the core there is no reason to keep the data
		// stored in memory without risking it staying unused till the program is terminated
		deleted, _ := clusterWrapper.DeleteSupervisor(supervisorInstance.Id)
		if deleted {
			thread.logger.Printf("deleted supervisor %d\n", supervisorInstance.Id)
		} else {
			thread.logger.Warnf("failed to delete supervisor %d\n", supervisorInstance.Id)
		}

		// let the modules thread decrement the semaphore otherwise we will be stuck in deadlock waiting for
		// the provisioned cluster to complete before allowing the etl-threads to shut down
		//if !clusterWrapper.IsStream() {
		thread.DecrementActiveSupervisors()
		thread.requestWg.Done()

		// if the processor is running in standalone mode there are two forms of compute that can exist:
		//	1. stream clusters that will run till the processes is asked to terminate with SIGINT
		//		~ the # of stream clusters present is represented by numOfActiveSupervisors
		//	2. clusters invoked through the commandline which would have completed at this point
		//
		// if we are in standalone mode and there are no active supervisors, no additional compute
		// will be performed on the processor, and we can shut down the threads
		if thread.Config.Standalone && (thread.numOfActiveSupervisors == 0) {
			thread.Interrupt <- threads.Shutdown
		}
	}(supervisorInstance)

	return nil
}
