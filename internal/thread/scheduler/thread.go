package scheduler

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/job"
	"github.com/GabeCordo/cluster-tools/internal/processor"
	scheduler "github.com/GabeCordo/cluster-tools/internal/scheduler/job"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/toolchain/multithreaded"
)

func (t *Thread) Setup() {

	var err error
	if t.Scheduler, err = scheduler.New(t.jobDatabase); err != nil {
		panic(err)
	}

	if err = t.Scheduler.Jobs.Load(t.config.SchedulesFolder); err != nil {
		panic(err)
	}
}

func (t *Thread) Start() {

	// RESPONSE THREADS

	go func() {
		for request := range t.C20 {
			t.wg.Add(1)
			t.ProcessRequest(&request)
			t.wg.Done()
		}
	}()

	go func() {
		// response coming from the processor thread
		for response := range t.C19 {
			t.processorResponseTable.Write(response.Nonce, response)
		}
	}()

	go func() {
		// response coming from the processor thread
		for response := range t.C27 {
			t.databaseResponseTable.Write(response.Nonce, response)
		}
	}()

	// SCHEDULER THREADS

	go scheduler.Watch(t.Scheduler)

	go scheduler.Loop(t.Scheduler, func(jb job.Job) error {

		// will return have a maximum of Timeout, so worst-case takes thread.config.Timeout
		mandatory := thread.Mandatory{t.C18, t.processorResponseTable, t.config.Timeout}
		_, err := thread.CreateSupervisor(mandatory, jb.Module, jb.Cluster, jb.Config, jb.Metadata)

		e := ""
		if err != nil {
			e = fmt.Sprintf("but encountered an error, %s", err.Error())
		}

		if (err != nil) && t.config.Debug {
			t.logger.Printf("scheduled cluster is ready: %s (%s,%s,%s) %s\n", jb.Identifier, jb.Module, jb.Cluster, jb.Config, e)
			t.logger.Printf("%d clusters are waiting to be provisioned\n", t.Scheduler.ItemsInQueue())
		}

		// if err is not nil, the Scheduler will stop running, so output to console
		// if debug is enabled so the operator is aware of the runtime change
		if ((err == processor.CanNotProvisionStreamCluster) || (err == multithreaded.NoResponseReceived)) && t.config.Debug {
			t.logger.Printf("the Scheduler stopped after encountering %s\n", err.Error())
		}

		// I only care about errors that might indicate a compromised state of the thread, the others
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

func (t *Thread) ProcessRequest(request *thread.Request) {

	response := thread.Response{Nonce: request.Nonce}

	switch request.Action {
	case thread.GetAction:
		switch request.Type {
		case thread.JobRecord:
			if filter, ok := (request.Data).(database.Filter); ok {
				response.Data = t.get(filter)
			} else {
				response.Success = false
				response.Error = thread.BadRequestType
			}
		case thread.QueueRecord:
			response.Data = t.queue()
			response.Success = true
		}

	case thread.CreateAction:
		if jb, ok := (request.Data).(job.Job); ok {
			response.Error = t.create(&jb)
			response.Success = response.Error == nil
			t.logger.Printf("created job:%s\n", jb.Identifier)
		} else {
			response.Success = false
			response.Error = thread.BadRequestType
		}
	case thread.DeleteAction:
		if filter, ok := (request.Data).(database.Filter); ok {
			response.Error = t.delete(filter)
			response.Success = response.Error == nil
			t.logger.Printf("deleted job:%s\n", filter.Identifier)
		} else {
			response.Success = false
			response.Error = thread.BadResponseType
		}
	}

	t.C21 <- response
}

func (t *Thread) Teardown() {

	// do not complete teardown until all requests have been completed
	t.wg.Wait()

	if db, ok := (t.Scheduler.Jobs).(database.Database); ok {

		if err := db.Save(t.config.SchedulesFolder); err != nil {
			t.logger.Panicln(err.Error())
		}
	}
}
