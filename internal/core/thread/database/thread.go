package database

import (
	"log"
	"time"

	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/statistic"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {

	if err := t.pipelineDatabase.Load(t.config.ConfigsFolder); err != nil {
		log.Panicf("could not load saved configs, statistic 'etl doctor' to verify the configuration is valid %s\n",
			err.Error())
	}

	// some configs may have carried over from previous runs
	// let the operator know these configs are being loaded into the
	// flock without having to query the database over HTTP
	t.pipelineDatabase.Print()
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
					if configData, ok := (request.Data).(pipeline.Pipeline); ok {
						_, err := t.pipelineDatabase.Create(
							database.Filter{
								Namespace: request.Identifiers.Namespace,
								Pipeline:  request.Identifiers.Pipeline,
							},
							&configData,
						)

						if err == nil {
							t.pipelineDatabase.Print()
						}

						response.Success = err == nil
					} else {
						response.Success = false
						response.Error = StoreTypeMismatch
					}
				}
			case thread.StatisticRecord:
				{
					if statisticsData, ok := (request.Data).(*statistic.Statistics); ok {
						_, err := t.statisticDatabase.Create(
							database.Filter{
								Namespace: request.Identifiers.Namespace,
								Pipeline:  request.Identifiers.Pipeline,
							},
							statistic.Wrapper{ // TODO : depreciate or fix elapsed time
								Timestamp: time.Now(),
								Stats:     *statisticsData, // copy
							},
						)

						if err == nil {
							t.statisticDatabase.Print()
						}
						response.Success = err == nil
					} else {
						response.Success = false
						response.Error = StoreTypeMismatch
					}
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
					results := t.pipelineDatabase.Get(database.Filter{
						Namespace:  request.Identifiers.Namespace,
						Identifier: request.Identifiers.Pipeline,
					})

					configs := make([]pipeline.Pipeline, len(results))
					for i, result := range results {
						configs[i] = result.(pipeline.Pipeline)
					}

					response.Success = len(results) > 0
					response.Data = configs
				}
			case thread.StatisticRecord:
				{
					results := t.statisticDatabase.Get(database.Filter{
						Namespace: request.Identifiers.Namespace,
						Pipeline:  request.Identifiers.Pipeline,
					})

					statistics := make([]statistic.Statistics, len(results))
					for i, result := range results {
						statistics[i] = result.(statistic.Statistics)
					}

					response.Success = len(results) > 0
					response.Data = statistics
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
					err := t.pipelineDatabase.Delete(database.Filter{
						Namespace:  request.Identifiers.Namespace,
						Identifier: request.Identifiers.Pipeline,
					})

					if db, ok := (t.pipelineDatabase).(database.Database); (err == nil) && ok {
						db.Print()
					}

					response.Success = err == nil
				}
			case thread.StatisticRecord:
				{
					err := t.statisticDatabase.Delete(database.Filter{
						Namespace: request.Identifiers.Namespace,
					})

					if db, ok := (t.statisticDatabase).(database.Database); (err == nil) && ok {
						db.Print()
					}

					response.Success = err == nil
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
					cfg := (request.Data).(pipeline.Pipeline)
					err := t.pipelineDatabase.Replace(database.Filter{
						Namespace: request.Identifiers.Namespace,
						Pipeline:  request.Identifiers.Pipeline,
					}, &cfg)

					if db, ok := (t.pipelineDatabase).(database.Database); (err == nil) && ok {
						db.Print()
					}

					response.Success = err == nil
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

	if err := t.pipelineDatabase.Save(t.config.ConfigsFolder); err != nil {
		log.Printf("failed to save configs created during runtime %s\n", err.Error())
	}

	if err := t.statisticDatabase.Save(t.config.StatisticsFolder); err != nil {
		log.Printf("failed to save statistics created during runtime %s\n", err.Error())
	}
}
