package scheduler

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/core/components/processor"
	"github.com/GabeCordo/cluster-tools/core/components/scheduler"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/multithreaded"
)

func (thread *Thread) Setup() {

	var err error
	if thread.Scheduler, err = scheduler.New(); err != nil {
		panic(err)
	}

	if err = scheduler.Load(thread.Scheduler, common.DefaultSchedulesFolder); err != nil {
		panic(err)
	}
}

func (thread *Thread) Start() {

	// RESPONSE THREADS

	go func() {
		for request := range thread.C20 {
			thread.wg.Add(1)
			thread.ProcessRequest(&request)
			thread.wg.Done()
		}
	}()

	go func() {
		// response coming from the processor thread
		for response := range thread.C19 {
			thread.responseTable.Write(response.Nonce, response)
		}
	}()

	// SCHEDULER THREADS

	go scheduler.Watch(thread.Scheduler)

	go scheduler.Loop(thread.Scheduler, func(job *scheduler.Job) error {

		// will return have a maximum of Timeout, so worst-case takes thread.config.Timeout
		mandatory := common.ThreadMandatory{thread.C18, thread.responseTable, thread.config.Timeout}
		_, err := common.CreateSupervisor(mandatory, job.Module, job.Cluster, job.Config, job.Metadata)

		e := ""
		if err != nil {
			e = fmt.Sprintf("but encountered an error, %s", err.Error())
		}

		if (err != nil) && thread.config.Debug {
			thread.logger.Printf("scheduled cluster is ready: %s (%s,%s,%s) %s\n", job.Identifier, job.Module, job.Cluster, job.Config, e)
			thread.logger.Printf("%d clusters are waiting to be provisioned\n", thread.Scheduler.ItemsInQueue())
		}

		// if err is not nil, the Scheduler will stop running, so output to console
		// if debug is enabled so the operator is aware of the runtime change
		if ((err == processor.CanNotProvisionStreamCluster) || (err == multithreaded.NoResponseReceived)) && thread.config.Debug {
			thread.logger.Printf("the Scheduler stopped after encountering %s\n", err.Error())
		}

		// I only care about errors that might indicate a compromised state of the threads, the others
		// like Module/Cluster's not existing really makes no sense to crash the Scheduler as someone
		// likely put in the job for a future module/cluster pair they want to attach to mango
		if (err == processor.CanNotProvisionStreamCluster) || (err == multithreaded.NoResponseReceived) ||
			(err == processor.ModuleDoesNotExist) || (err == processor.ClusterDoesNotExist) {
			return err
		} else {
			return nil
		}
	})
}

func (thread *Thread) ProcessRequest(request *common.ThreadRequest) {

	response := common.ThreadResponse{Nonce: request.Nonce}

	switch request.Action {
	case common.GetAction:
		switch request.Type {
		case common.JobRecord:
			if filter, ok := (request.Data).(scheduler.Filter); ok {
				response.Data = thread.get(&filter)
			} else {
				response.Success = false
				response.Error = common.BadRequestType
			}
		case common.QueueRecord:
			response.Data = thread.queue()
			response.Success = true
		}

	case common.CreateAction:
		if job, ok := (request.Data).(scheduler.Job); ok {
			response.Error = thread.create(&job)
			response.Success = response.Error == nil
		} else {
			response.Success = false
			response.Error = common.BadRequestType
		}
	case common.DeleteAction:
		if filter, ok := (request.Data).(scheduler.Filter); ok {
			response.Error = thread.delete(&filter)
			response.Success = response.Error == nil
		} else {
			response.Success = false
			response.Error = common.BadResponseType
		}
	}

	thread.C21 <- response
}

func (thread *Thread) Teardown() {

	// do not complete teardown until all requests have been completed
	thread.wg.Wait()

	if err := scheduler.Save(thread.Scheduler, common.DefaultSchedulesFolder); err != nil {
		thread.logger.Panicln(err.Error())
	}

}
