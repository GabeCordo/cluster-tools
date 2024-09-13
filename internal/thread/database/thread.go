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

	go func() {
		// request from http_server
		for request := range t.C1 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			request.Source = thread.HttpClient
			t.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range t.C11 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			request.Source = thread.Processor
			t.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range t.C15 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			request.Source = thread.Supervisor
			t.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from scheduler
		for request := range t.C26 {
			if !t.accepting {
				break
			}
			t.wg.Add(1)

			request.Source = thread.Supervisor
			t.ProcessIncomingRequest(&request)
		}
	}()

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

func (t *Thread) Respond(request *thread.Request, response *thread.Response) (success bool) {

	success = true

	switch request.Source {
	case thread.HttpClient:
		t.C2 <- *response
		break
	case thread.Processor:
		t.C12 <- *response
		break
	case thread.Supervisor:
		t.C16 <- *response
		break
	case thread.Scheduler:
		t.C27 <- *response
		break
	default:
		success = false
	}

	return success
}

func (t *Thread) ProcessIncomingRequest(request *thread.Request) {

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

						t.Respond(request, &thread.Response{
							Success: err == nil,
							Nonce:   request.Nonce,
						})
					} else {
						t.Respond(request, &thread.Response{
							Success: false,
							Nonce:   request.Nonce,
							Error:   StoreTypeMismatch,
						})
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

						if db, ok := (t.statisticDatabase).(database.Database); (err == nil) && ok {
							db.Print()
						}

						t.Respond(request, &thread.Response{
							Success: err == nil,
							Nonce:   request.Nonce,
						})
					} else {
						t.Respond(request, &thread.Response{
							Success: false,
							Nonce:   request.Nonce,
							Error:   StoreTypeMismatch,
						})
					}
				}
			case thread.JobRecord:
				{

				}
			}
		}
	case thread.GetAction:
		{
			var response thread.Response

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

					response = thread.Response{Success: len(results) > 0, Nonce: request.Nonce, Data: configs}
					t.Respond(request, &response)
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

					response = thread.Response{Success: len(results) > 0, Nonce: request.Nonce, Data: statistics}
					t.Respond(request, &response)
				}
			case thread.JobRecord:
				{

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

					response := thread.Response{Success: err == nil, Nonce: request.Nonce}
					t.Respond(request, &response)
				}
			case thread.StatisticRecord:
				{
					err := t.statisticDatabase.Delete(database.Filter{Module: request.Identifiers.Module})

					if db, ok := (t.statisticDatabase).(database.Database); (err == nil) && ok {
						db.Print()
					}

					response := thread.Response{Success: err == nil, Nonce: request.Nonce}
					t.Respond(request, &response)
				}
			case thread.JobRecord:
				{

				}
			}
		}
	case thread.UpdateAction:
		{
			switch request.Type {
			case thread.ConfigRecord:
				cfg := (request.Data).(config.Config)
				err := t.configDatabase.Replace(database.Filter{Module: request.Identifiers.Module, Cluster: request.Identifiers.Cluster}, &cfg)

				if db, ok := (t.configDatabase).(database.Database); (err == nil) && ok {
					db.Print()
				}

				response := thread.Response{Success: err == nil, Nonce: request.Nonce}
				t.Respond(request, &response)
			}
		}
	}

	t.wg.Done()
}

func (t *Thread) ProcessIncomingResponse(response *thread.Response) {
	t.messengerResponseTable.Write(response.Nonce, response)
}
