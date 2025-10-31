package scheduler

import (
	"github.com/FortifiedCode/flock/internal/core/thread"
)

func (t *Thread) Setup() {

}

func (t *Thread) Start() {

	go t.useCases.SchedulerWatch()

	go t.useCases.SchedulerLoop(func(namespaceId, pipelineId string, metadata map[string]string) error {
		// will return have a maximum of Timeout, so worst-case takes thread.pipeline.Timeout
		mandatory := thread.Mandatory{
			Pipe:          t.channels.c18,
			ResponseTable: t.processorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		}
		_, err := thread.CreateRun(mandatory, namespaceId, pipelineId, metadata)
		return err
	})

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
					t.handleGetJob(request, response)
				}
			case thread.QueueRecord:
				{
					t.handleGetQueue(request, response)
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
			t.handleCreateJob(request, response)
		}
	case thread.DeleteAction:
		{
			t.handleDeleteJob(request, response)
		}
	default:
		{
			response.Success = false
			response.Error = thread.UnknownRequest
		}
	}

	return response
}

func (t *Thread) TearDown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown

	t.useCases.SaveJobsToDisk(t.config.SchedulesFolder)
}
