package database

import (
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/config"
	"github.com/GabeCordo/cluster-tools/internal/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"log"
	"time"
)

func (t *Thread) Setup() {
	t.accepting = true

	if err := t.configDatabase.Load(t.config.ConfigsFolder); err != nil {
		log.Panicf("could not load saved configs, run 'etl doctor' to verify the configuration is valid %s\n",
			err.Error())
	}

	// some configs may have carried over from previous runs
	// let the operator know these configs are being loaded into the
	// cluster-tools without having to query the database over HTTP
	t.configDatabase.Print()
}

func (t *Thread) Teardown() {
	t.accepting = false

	if err := t.configDatabase.Save(t.config.ConfigsFolder); err != nil {
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
			case thread.ConfigRecord:
				{
					if configData, ok := (request.Data).(config.Config); ok {
						_, err := t.configDatabase.Create(
							database.Filter{
								Module:  request.Identifiers.Module,
								Cluster: request.Identifiers.Cluster,
							},
							&configData,
						)

						if err == nil {
							t.configDatabase.Print()
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
								Module:  request.Identifiers.Module,
								Cluster: request.Identifiers.Cluster,
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
			case thread.ConfigRecord:
				{
					results := t.configDatabase.Get(database.Filter{
						Module:     request.Identifiers.Module,
						Identifier: request.Identifiers.Cluster,
					})

					configs := make([]config.Config, len(results))
					for i, result := range results {
						configs[i] = result.(config.Config)
					}

					response.Success = len(results) > 0
					response.Data = configs
				}
			case thread.StatisticRecord:
				{
					results := t.statisticDatabase.Get(database.Filter{
						Module:  request.Identifiers.Module,
						Cluster: request.Identifiers.Cluster,
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
					t.logger.Warn(thread.UnknownRequest.Error())
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.ConfigRecord:
				{
					err := t.configDatabase.Delete(database.Filter{Module: request.Identifiers.Module, Identifier: request.Identifiers.Config})

					if db, ok := (t.configDatabase).(database.Database); (err == nil) && ok {
						db.Print()
					}

					response.Success = err == nil
				}
			case thread.StatisticRecord:
				{
					err := t.statisticDatabase.Delete(database.Filter{Module: request.Identifiers.Module})

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
			case thread.ConfigRecord:
				{
					cfg := (request.Data).(config.Config)
					err := t.configDatabase.Replace(database.Filter{Module: request.Identifiers.Module, Cluster: request.Identifiers.Cluster}, &cfg)

					if db, ok := (t.configDatabase).(database.Database); (err == nil) && ok {
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
