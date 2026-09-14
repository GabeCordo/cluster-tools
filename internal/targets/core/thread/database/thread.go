package database

import (
	"github.com/GabeCordo/DistributedFunctions/internal/flags"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/terminal"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/thread"
)

func (t *Thread) Setup() {

	t.logger.SetColour(terminal.Blue)

	err := t.useCases.LoadDatabases(t.config.ConfigsFolder)
	if err != nil {
		panic(err)
	}
}

func (t *Thread) Start() {

	var iReq *thread.Request
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-t.channels.c1:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c2 <- oRsp
				}
			}
		case iReq = <-t.channels.c11:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c12 <- oRsp
				}
			}
		case iReq = <-t.channels.c15:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c16 <- oRsp
				}
			}
		case iReq = <-t.channels.c26:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c27 <- oRsp
				}
			}
		case <-t.channels.close:
			{
				// shutting down the database thread
				break
			}
		}
		oRsp = nil
	}
}

func (t *Thread) handleRequest(request *thread.Request) (response *thread.Response) {

	if flags.DEBUG {
		t.logger.Printf("Received %s", request.ToString())
	}

	response = thread.NewResponse(thread.Database)

	switch request.Action {
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.PipelineRecord:
				{
					t.handleCreatePipelineRecord(request, response)
				}
			case thread.StatisticRecord:
				{
					t.handleCreateStatisticRecord(request, response)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.GetAction:
		{
			switch request.Type {
			case thread.PipelineRecord:
				{
					t.handleGetPipelineRecord(request, response)
				}
			case thread.StatisticRecord:
				{
					t.handleGetStatisticRecord(request, response)
				}
			case thread.NamespaceRecord:
				{
					t.handleGetNamespaceRecord(request, response)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.SummaryAction:
		{
			switch request.Type {
			case thread.StatisticRecord:
				{
					t.handleSummaryStatistics(request, response)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.PipelineRecord:
				{
					t.handleDeletePipelineRecord(request, response)
				}
			case thread.StatisticRecord:
				{
					t.handleDeleteStatisticRecord(request, response)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	case thread.UpdateAction:
		{
			switch request.Type {
			case thread.PipelineRecord:
				{
					t.handleUpdatePipelineRecord(request, response)
				}
			default:
				{
					response.Error = thread.UnknownRequest
				}
			}
		}
	default:
		{
			response.Error = thread.UnknownRequest
		}
	}

	response.Success = response.Error == nil
	return response
}

func (t *Thread) Teardown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
