package database

import (
	"github.com/Sentmint/pops/internal/core/database"
	"github.com/Sentmint/pops/internal/core/database/statistic"
	"github.com/Sentmint/pops/internal/core/thread"
	"github.com/Sentmint/yule"
	"log"
	"time"
)

func (t *Thread) Setup() {
	t.accepting = true

	if err := t.pipelineDatabase.Load(t.config.ConfigsFolder); err != nil {
		log.Panicf("could not load saved configs, statistic 'etl doctor' to verify the configuration is valid %s\n",
			err.Error())
	}

	// some configs may have carried over from previous runs
	// let the operator know these configs are being loaded into the
	// pops-core without having to query the database over HTTP
	t.pipelineDatabase.Print()
}

func (t *Thread) Teardown() {
	t.accepting = false

	if err := t.pipelineDatabase.Save(t.config.ConfigsFolder); err != nil {
		log.Printf("failed to save configs created during runtime %s\n", err.Error())
	}

	if err := t.statisticDatabase.Save(t.config.StatisticsFolder); err != nil {
		log.Printf("failed to save statistics created during runtime %s\n", err.Error())
	}

	t.wg.Wait()
}

func (t *Thread) Start() {

	// LISTEN FOR INCOMING REQUESTS

	thread.SetupListener(t.C1, t.C2, &t.accepting, &t.wg, thread.Database, t.Handle)

	thread.SetupListener(t.C11, t.C12, &t.accepting, &t.wg, thread.Database, t.Handle)

	thread.SetupListener(t.C15, t.C16, &t.accepting, &t.wg, thread.Database, t.Handle)

	thread.SetupListener(t.C26, t.C27, &t.accepting, &t.wg, thread.Database, t.Handle)

	// LISTEN FOR INCOMING RESPONSES

	go func() {
		for response := range t.C4 {
			if !t.accepting {
				break
			}
			t.ProcessIncomingResponse(&response)
		}
	}()
}

func (t *Thread) Request(module thread.Module, request any) (success bool) {

	success = true

	switch module {
	case thread.Messenger:
		t.C3 <- *(request).(*thread.Request)
	default:
		success = false
	}
	return success
}

func (t *Thread) Handle(request *thread.Request, response *thread.Response) {

	switch request.Action {
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.PipelineRecord:
				{
					if configData, ok := (request.Data).(yule.Pipeline); ok {
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
					if statisticsData, ok := (request.Data).(*yule.Statistics); ok {
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
					t.logger.Warn(thread.UnknownRequest.Error())
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

					configs := make([]yule.Pipeline, len(results))
					for i, result := range results {
						configs[i] = result.(yule.Pipeline)
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

					statistics := make([]yule.Statistics, len(results))
					for i, result := range results {
						statistics[i] = result.(yule.Statistics)
					}

					response.Success = len(results) > 0
					response.Data = statistics
				}
			default:
				{
					t.logger.Warn(thread.UnknownRequest.Error())
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
					t.logger.Warn(thread.UnknownRequest.Error())
				}
			}
		}
	case thread.UpdateAction:
		{
			switch request.Type {
			case thread.PipelineRecord:
				{
					cfg := (request.Data).(yule.Pipeline)
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
					t.logger.Warn(thread.UnknownRequest.Error())
				}
			}
		}
	default:
		t.logger.Warn(thread.UnknownRequest.Error())
	}
}

func (t *Thread) ProcessIncomingResponse(response *thread.Response) {
	t.messengerResponseTable.Write(response.Nonce, response)
}
