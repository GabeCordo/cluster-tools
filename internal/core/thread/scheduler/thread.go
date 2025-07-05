package scheduler

import (
	job2 "github.com/GabeCordo/Flock/internal/core/component/scheduler/job"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/job"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {

	var err error
	if t.Scheduler, err = job2.New(t.jobDatabase); err != nil {
		panic(err)
	}

	if err = t.Scheduler.Jobs.Load(t.config.SchedulesFolder); err != nil {
		panic(err)
	}
}

func (t *Thread) Start() {

	go t.watch()
	go t.loop()

	var iReq *thread.Request
	var iRsp *thread.Response
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-t.channels.c20:
			{
				oRsp = t.HandleRequest(iReq)
				if oRsp != nil {
					t.channels.c21 <- oRsp
				}
			}
		case iRsp = <-t.channels.c19:
			{
				// response coming from the processor thread
				t.processorResponseTable.Write(iRsp.Nonce, iRsp)
			}
		case iRsp = <-t.channels.c27:
			{
				// response coming from the database thread
				t.databaseResponseTable.Write(iRsp.Nonce, iRsp)
			}
		case <-t.channels.close:
			{
				// shutting down the scheduler thread
				break
			}
		}
		oRsp = nil
	}
}

func (t *Thread) HandleRequest(request *thread.Request) (response *thread.Response) {

	response = thread.NewResponse(thread.Scheduler)
	thread.CopyMetadata(request, response)

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
					response.Success = false
					response.Error = thread.UnknownRequest
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
			response.Success = false
			response.Error = thread.UnknownRequest
		}
	}

	return response
}

func (t *Thread) Teardown() {

	// do not complete teardown until all requests have been completed
	t.wg.Wait()

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown

	if db, ok := (t.Scheduler.Jobs).(database.Database); ok {

		if err := db.Save(t.config.SchedulesFolder); err != nil {
			t.logger.Panicln(err.Error())
		}
	}
}
