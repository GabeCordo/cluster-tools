package database

import (
	"github.com/GabeCordo/cluster-tools/core/components/database"
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"log"
	"time"
)

func (thread *Thread) Setup() {
	thread.accepting = true

	thread.configFolderPath = common.DefaultConfigsFolder
	thread.statisticFolderPath = common.DefaultStatisticsFolder

	if err := GetConfigDatabaseInstance().Load(thread.configFolderPath); err != nil {
		log.Panicf("could not load saved configs, run 'etl doctor' to verify the configuration is valid %s\n",
			err.Error())
	}

	// some configs may have carried over from previous runs
	// let the operator know these configs are being loaded into the
	// core without having to query the database over HTTP
	GetConfigDatabaseInstance().Print()
}

func (thread *Thread) Teardown() {
	thread.accepting = false

	if err := GetConfigDatabaseInstance().Save(thread.configFolderPath); err != nil {
		log.Printf("failed to save configs created during runtime %s\n", err.Error())
	}

	if err := GetStatisticDatabaseInstance().Save(thread.statisticFolderPath); err != nil {
		log.Printf("failed to save statistics created during runtime %s\n", err.Error())
	}

	thread.wg.Wait()
}

func (thread *Thread) Start() {

	// LISTEN FOR INCOMING REQUESTS

	go func() {
		// request from http_server
		for request := range thread.C1 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.HttpClient
			thread.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range thread.C11 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.Processor
			thread.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range thread.C15 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.Supervisor
			thread.ProcessIncomingRequest(&request)
		}
	}()

	// LISTEN FOR INCOMING RESPONSES

	go func() {
		for response := range thread.C4 {
			if !thread.accepting {
				break
			}
			thread.ProcessIncomingResponse(&response)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) Request(module common.Module, request any) (success bool) {

	success = true

	switch module {
	case common.Messenger:
		thread.C3 <- *(request).(*common.ThreadRequest)
	default:
		success = false
	}
	return success
}

func (thread *Thread) Respond(request *common.ThreadRequest, response *common.ThreadResponse) (success bool) {

	success = true

	switch request.Source {
	case common.HttpClient:
		thread.C2 <- *response
		break
	case common.Processor:
		thread.C12 <- *response
		break
	case common.Supervisor:
		thread.C16 <- *response
	default:
		success = false
	}

	return success
}

func (thread *Thread) ProcessIncomingRequest(request *common.ThreadRequest) {

	switch request.Action {
	case common.CreateAction:
		{
			switch request.Type {
			case common.ConfigRecord:
				{
					if configData, ok := (request.Data).(interfaces.Config); ok {
						err := GetConfigDatabaseInstance().Create(request.Identifiers.Module, request.Identifiers.Cluster, configData)

						if err == nil {
							GetConfigDatabaseInstance().Print()
						}

						thread.Respond(request, &common.ThreadResponse{
							Success: err == nil,
							Nonce:   request.Nonce,
						})
					} else {
						thread.Respond(request, &common.ThreadResponse{
							Success: false,
							Nonce:   request.Nonce,
							Error:   StoreTypeMismatch,
						})
					}
				}
			case common.StatisticRecord:
				{
					if statisticsData, ok := (request.Data).(*interfaces.Statistics); ok {
						err := GetStatisticDatabaseInstance().Create(
							request.Identifiers.Module, request.Identifiers.Cluster,
							database.Statistic{ // TODO : depreciate or fix elapsed time
								Timestamp: time.Now(),
								Stats:     *statisticsData, // copy
							})

						if err == nil {
							GetStatisticDatabaseInstance().Print()
						}

						thread.Respond(request, &common.ThreadResponse{
							Success: err == nil,
							Nonce:   request.Nonce,
						})
					} else {
						thread.Respond(request, &common.ThreadResponse{
							Success: false,
							Nonce:   request.Nonce,
							Error:   StoreTypeMismatch,
						})
					}
				}
			}
		}
	case common.GetAction:
		{
			var response common.ThreadResponse

			switch request.Type {
			case common.ConfigRecord:
				{
					config, err := GetConfigDatabaseInstance().Get(database.ConfigFilter{
						Module:     request.Identifiers.Module,
						Identifier: request.Identifiers.Cluster,
					})
					if err != nil {
						response = common.ThreadResponse{Success: false, Nonce: request.Nonce}
					} else {
						// TODO - use to expect one record, now will have many
						response = common.ThreadResponse{Success: true, Nonce: request.Nonce, Data: config}
					}

					thread.Respond(request, &response)
				}
			case common.StatisticRecord:
				{
					records, err := GetStatisticDatabaseInstance().Get(database.StatisticFilter{
						Module:  request.Identifiers.Module,
						Cluster: request.Identifiers.Cluster,
					})
					response = common.ThreadResponse{Success: err == nil, Nonce: request.Nonce, Data: records}
					thread.Respond(request, &response)
				}
			}
		}
	case common.DeleteAction:
		{
			switch request.Type {
			case common.ConfigRecord:
				{
					err := GetConfigDatabaseInstance().Delete(request.Identifiers.Module, request.Identifiers.Cluster)

					if err == nil {
						GetConfigDatabaseInstance().Print()
					}

					response := common.ThreadResponse{Success: err == nil, Nonce: request.Nonce}
					thread.Respond(request, &response)
				}
			case common.StatisticRecord:
				{
					err := GetStatisticDatabaseInstance().Delete(request.Identifiers.Module)

					if err == nil {
						GetConfigDatabaseInstance().Print()
					}

					response := common.ThreadResponse{Success: err == nil, Nonce: request.Nonce}
					thread.Respond(request, &response)
				}
			}
		}
	case common.UpdateAction:
		{
			switch request.Type {
			case common.ConfigRecord:
				config := (request.Data).(interfaces.Config)
				err := GetConfigDatabaseInstance().Replace(request.Identifiers.Module, request.Identifiers.Cluster, config)

				if err == nil {
					GetConfigDatabaseInstance().Print()
				}

				response := common.ThreadResponse{Success: err == nil, Nonce: request.Nonce}
				thread.Respond(request, &response)
			}
		}
	}

	thread.wg.Done()
}

func (thread *Thread) ProcessIncomingResponse(response *common.ThreadResponse) {
	thread.messengerResponseTable.Write(response.Nonce, response)
}
