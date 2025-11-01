package scheduler

import (
	thread2 "github.com/FortifiedCode/flock/internal/targets/core/thread"
)

func (t *Thread) Setup() {

}

func (t *Thread) Start() {

	go t.useCases.SchedulerWatch()

	go t.useCases.SchedulerLoop(func(namespaceId, pipelineId string, metadata map[string]string) error {
		// will return have a maximum of Timeout, so worst-case takes thread.pipeline.Timeout
		mandatory := thread2.Mandatory{
			Pipe:          t.channels.c18,
			ResponseTable: t.processorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		}
		_, err := thread2.CreateRun(mandatory, namespaceId, pipelineId, metadata)
		return err
	})

	var iReq *thread2.Request
	var iRsp *thread2.Response
	var oRsp *thread2.Response

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

func (t *Thread) HandleRequest(request *thread2.Request) (response *thread2.Response) {

	response = thread2.NewResponse(thread2.Scheduler)
	thread2.CopyMetadata(request, response)

	switch request.Action {
	case thread2.GetAction:
		{
			switch request.Type {
			case thread2.JobRecord:
				{
					t.handleGetJob(request, response)
				}
			case thread2.QueueRecord:
				{
					t.handleGetQueue(request, response)
				}
			default:
				{
					response.Success = false
					response.Error = thread2.UnknownRequest
				}
			}
		}
	case thread2.CreateAction:
		{
			t.handleCreateJob(request, response)
		}
	case thread2.DeleteAction:
		{
			t.handleDeleteJob(request, response)
		}
	default:
		{
			response.Success = false
			response.Error = thread2.UnknownRequest
		}
	}

	return response
}

func (t *Thread) TearDown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread2.Shutdown

	t.useCases.SaveJobsToDisk(t.config.SchedulesFolder)
}
