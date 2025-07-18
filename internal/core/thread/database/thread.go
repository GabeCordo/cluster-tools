package database

import (
	"github.com/GabeCordo/Flock/internal/core/thread"
	"log"
)

func (t *Thread) Setup() {

	if err := t.useCases.PipelineDatabase.Load(t.config.ConfigsFolder); err != nil {
		log.Panicf("could not load saved configs, statistic 'etl doctor' to verify the configuration is valid %s\n",
			err.Error())
	}

	// some configs may have carried over from previous runs
	// let the operator know these configs are being loaded into the
	// flock without having to query the database over HTTP
	t.useCases.PipelineDatabase.Print()
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

	t.wg.Wait()

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown

	if err := t.useCases.PipelineDatabase.Save(t.config.ConfigsFolder); err != nil {
		log.Printf("failed to save configs created during runtime %s\n", err.Error())
	}

	if err := t.useCases.StatisticDatabase.Save(t.config.StatisticsFolder); err != nil {
		log.Printf("failed to save statistics created during runtime %s\n", err.Error())
	}
}
