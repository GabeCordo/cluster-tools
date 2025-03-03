package provisioner

import (
	"errors"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/Sentmint/pops/internal/core/database/run"
	"github.com/Sentmint/pops/internal/processor/threads"
	"github.com/Sentmint/yule"
	"math/rand"
	"sync"
	"time"
)

func (thread *Thread) getSupervisor() []*yule.Pipeline {

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

	pInstance := yule.Build(request.Pipeline, thread.repository)

	// TODO : fix
	//if err != nil {
	//	thread.logger.Printf("%s[%s]%s Failed to create runner %s\n", logging.Red, request.Pipeline.Identifier, logging.Reset, err)
	//	thread.requestWg.Done()
	//	return err
	//}

	thread.logger.Printf("%s[%s]%s Pipeline Active (run: %d)\n", logging.Green, request.Pipeline.Identifier, logging.Reset, pInstance.Identifier)
	go func(p yule.RunnablePipeline) {

		m := sync.Mutex{} // used for sending updates to the gateway

		// premise:
		// the statistics associated with the running pipeline instance will be updated on-demand
		// as data flows through the channels between functions.
		//
		// idea:
		// every 1s send an update of the statistics to the PipelineOps gateway so the operator
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
				if !p.IsAlive() {
					thread.logger.Warnf("cannot send update for supervisor %d that is not alive\n", pInstance.Identifier)
					break
				}

				run := run.Run{
					Id:         p.Identifier,
					Status:     run.Active,
					Statistics: p.Stats,
				}

				thread.C0 <- threads.SocketRequest{
					Action: threads.SocketRunUpdate,
					Data:   run,
					Nonce:  rand.Uint32(),
				}

				m.Unlock()

				time.Sleep(1 * time.Second) // wait before the next update
			}
		}()

		thread.runWg.Add(1)

		// block until the runner completes
		p.Run()

		m.Lock()

		run := run.Run{
			Id:         p.Identifier,
			Status:     run.FromString(p.GetState().ToString()), // TODO: provision.RunStatus(supervisorInstance.State)
			Statistics: p.Stats,
		}

		thread.C0 <- threads.SocketRequest{
			Action: threads.SocketRunUpdate,
			Data:   run,
			Nonce:  rand.Uint32(),
		}

		m.Unlock()

		// provide the console with output indicating that the cluster has completed
		// we already provide output when a cluster is provisioned, so it completes the state
		if *thread.Config.Debug {
			duration := time.Now().Sub(p.Stats.StartTime)
			thread.logger.Printf("%s[%d]%s Pipeline complete, took %dhr %dm %ds %dms %dus\n",
				logging.Green,
				p.Identifier,
				logging.Reset,
				int(duration.Hours()),
				int(duration.Minutes()),
				int(duration.Seconds()),
				int(duration.Milliseconds()),
				int(duration.Microseconds()),
			)
		}

		// TODO : do we need this anymore?
		// once a runner has sent its data to the core there is no reason to keep the data
		// stored in memory without risking it staying unused till the program is terminated
		//deleted, _ := thread.DeleteSupervisor(supervisorInstance.Id)
		//if deleted {
		//	thread.logger.Printf("deleted runner %d\n", p.Id)
		//} else {
		//	thread.logger.Warnf("failed to delete runner %d\n", p.Id)
		//}

		// let the modules thread decrement the semaphore otherwise we will be stuck in deadlock waiting for
		// the provisioned cluster to complete before allowing the etl-threads to shut down
		//if !clusterWrapper.IsStream() {
		thread.DecrementActiveSupervisors()
		thread.runWg.Done()
	}(pInstance)

	return nil
}

func (thread *Thread) stopRun(request *threads.ProvisionerRequest) error {

	thread.backlogMutex.Lock()
	defer thread.backlogMutex.Unlock()

	foundRunnable := false
	for _, r := range thread.runnablePipelines {
		if r.Identifier == request.Supervisor {
			foundRunnable = true
			r.Teardown()
			break
		}
	}

	if !foundRunnable {
		return errors.New("no supervisor with that id exists")
	} else {
		return nil
	}
}
