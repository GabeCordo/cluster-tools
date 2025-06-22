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

func (t *Thread) Setup() {

	var err error
	if t.Scheduler, err = job2.New(t.jobDatabase); err != nil {
		panic(err)
	}

	if err = t.Scheduler.Jobs.Load(t.config.SchedulesFolder); err != nil {
		panic(err)
	}

	t.accepting = true
}

func (t *Thread) Start() {

	// LISTENER THREADS

	thread.SetupListener(t.channels.c20, t.channels.c21, &t.accepting, &t.wg, thread.Scheduler, t.HandleRequest)

	// RESPONSE THREADS

	go func() {
		// response coming from the processor thread
		for response := range t.channels.c19 {
			t.processorResponseTable.Write(response.Nonce, response)
		}
	}()

	go func() {
		// response coming from the processor thread
		for response := range t.channels.c27 {
			t.databaseResponseTable.Write(response.Nonce, response)
		}
	}()

	// SCHEDULER THREADS

	go job2.Watch(t.Scheduler)

	go func() {
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
	}()
}

func (t *Thread) HandleRequest(request *thread.Request, response *thread.Response) {

	response.Type = request.Type
	response.Action = request.Action
	response.Nonce = request.Nonce

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.JobRecord:
				{
					if filter, ok := (request.Data).(database.Filter); ok {
						response.Data = t.get(filter)
					} else {
						response.Success = false
						response.Error = thread.BadRequestType
					}
				}
			case thread.QueueRecord:
				{
					response.Data = t.queue()
					response.Success = true
				}
			default:
				{
					t.logger.Warn(thread.UnknownRequest.Error())
				}
			}
		}
	case thread.CreateAction:
		{
			if jb, ok := (request.Data).(job.Job); ok {
				response.Error = t.create(&jb)
				response.Success = response.Error == nil
				t.logger.Printf("created job:%s\n", jb.Identifier)
			} else {
				response.Success = false
				response.Error = thread.BadRequestType
			}
		}
	case thread.DeleteAction:
		{
			if filter, ok := (request.Data).(database.Filter); ok {
				response.Error = t.delete(filter)
				response.Success = response.Error == nil
				t.logger.Printf("deleted job:%s\n", filter.Identifier)
			} else {
				response.Success = false
				response.Error = thread.BadResponseType
			}
		}
	default:
		{
			t.logger.Warn(thread.UnknownRequest.Error())
		}
	}
}

func (t *Thread) Teardown() {

	t.accepting = false

	// do not complete teardown until all requests have been completed
	t.wg.Wait()

	if db, ok := (t.Scheduler.Jobs).(database.Database); ok {

		if err := db.Save(t.config.SchedulesFolder); err != nil {
			t.logger.Panicln(err.Error())
		}
	}
}
