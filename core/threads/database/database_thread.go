package database

import (
	"github.com/GabeCordo/mango/core/components/database"
	"github.com/GabeCordo/mango/core/interfaces/cluster"
	"github.com/GabeCordo/mango/core/threads/common"
	"github.com/GabeCordo/toolchain/multithreaded"
	"log"
	"math/rand"
	"time"
)

func (thread *Thread) Setup() {
	thread.accepting = true

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

			request.Source = common.HttpProcessor
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
		thread.C3 <- *(request).(*common.MessengerRequest)
	default:
		success = false
	}
	return success
}

func (thread *Thread) Respond(request *common.DatabaseRequest, response *common.DatabaseResponse) (success bool) {

	success = true

	switch request.Source {
	case common.HttpClient:
		thread.C2 <- *response
		break
	case common.HttpProcessor:
		thread.C12 <- *response
		break
	case common.Supervisor:
		thread.C16 <- *response
	default:
		success = false
	}

	return success
}

func (thread *Thread) ProcessIncomingRequest(request *common.DatabaseRequest) {

	switch request.Action {
	case common.DatabaseStore:
		{
			switch request.Type {
			case common.ClusterConfig:
				{
					configData := (request.Data).(cluster.Config)
					err := GetConfigDatabaseInstance().Create(request.Module, request.Cluster, configData)

					if err == nil {
						GetConfigDatabaseInstance().Print()
					}

					thread.Respond(request, &common.DatabaseResponse{
						Success: err == nil,
						Nonce:   request.Nonce,
					})
				}
			case common.SupervisorStatistic:
				{
					statisticsData := (request.Data).(*cluster.Statistics)
					err := GetStatisticDatabaseInstance().Create(
						request.Module, request.Cluster,
						database.Statistic{ // TODO : depreciate or fix elapsed time
							Timestamp: time.Now(),
							Stats:     *statisticsData, // copy
						})

					if err == nil {
						GetStatisticDatabaseInstance().Print()
					}

					thread.Respond(request, &common.DatabaseResponse{
						Success: err == nil,
						Nonce:   request.Nonce,
					})
				}
			}
		}
	case common.DatabaseFetch:
		{
			var response common.DatabaseResponse

			switch request.Type {
			case common.ClusterConfig:
				{
					config, err := GetConfigDatabaseInstance().Get(database.ConfigFilter{
						Module:     request.Module,
						Identifier: request.Cluster,
					})
					if err != nil {
						response = common.DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						// TODO - use to expect one record, now will have many
						response = common.DatabaseResponse{Success: true, Nonce: request.Nonce, Data: config}
					}

					thread.Respond(request, &response)
				}
			case common.SupervisorStatistic:
				{
					records, err := GetStatisticDatabaseInstance().Get(database.StatisticFilter{
						Module:  request.Module,
						Cluster: request.Cluster,
					})
					response = common.DatabaseResponse{Success: err == nil, Nonce: request.Nonce, Data: records}
					thread.Respond(request, &response)
				}
			}
		}
	case common.DatabaseDelete:
		{
			switch request.Type {
			case common.ClusterConfig:
				{
					err := GetConfigDatabaseInstance().Delete(request.Module, request.Cluster)

					if err == nil {
						GetConfigDatabaseInstance().Print()
					}

					response := common.DatabaseResponse{Success: err == nil, Nonce: request.Nonce}
					thread.Respond(request, &response)
				}
			case common.ClusterModule:
				{
					err := GetStatisticDatabaseInstance().Delete(request.Module)

					if err == nil {
						GetConfigDatabaseInstance().Print()
					}

					response := common.DatabaseResponse{Success: err == nil, Nonce: request.Nonce}
					thread.Respond(request, &response)
				}
			}
		}
	case common.DatabaseReplace:
		{
			config := (request.Data).(cluster.Config)
			err := GetConfigDatabaseInstance().Replace(request.Module, request.Cluster, config)

			if err == nil {
				GetConfigDatabaseInstance().Print()
			}

			response := common.DatabaseResponse{Success: err == nil, Nonce: request.Nonce}
			thread.Respond(request, &response)
		}
	case common.DatabaseUpperPing:
		{
			thread.ProcessDatabaseUpperPing(request)
		}
	case common.DatabaseLowerPing:
		{
			thread.ProcessDatabaseLowerPing(request)
		}
	}

	thread.wg.Done()
}

func (thread *Thread) ProcessDatabaseUpperPing(request *common.DatabaseRequest) {

	if thread.config.Debug {
		thread.logger.Println("received ping over C1")
	}

	messengerPingRequest := &common.MessengerRequest{
		Action: common.MessengerUpperPing,
		Nonce:  rand.Uint32(),
	}
	thread.Request(common.Messenger, messengerPingRequest)

	data, didTimeout := multithreaded.SendAndWait(thread.messengerResponseTable, messengerPingRequest.Nonce,
		thread.config.MaxWaitForResponse)

	if didTimeout {
		thread.C2 <- common.DatabaseResponse{
			Nonce:   request.Nonce,
			Success: false,
		}
		return
	}

	messengerResponse := (data).(*common.MessengerResponse)

	if !messengerResponse.Success {
		thread.C2 <- common.DatabaseResponse{
			Nonce:   request.Nonce,
			Success: false,
		}
		return
	}

	if thread.config.Debug {
		thread.logger.Println("received ping over C4")
	}

	databaseResponse := common.DatabaseResponse{
		Nonce:   request.Nonce,
		Success: messengerResponse.Success,
	}
	thread.C2 <- databaseResponse
}

func (thread *Thread) ProcessDatabaseLowerPing(request *common.DatabaseRequest) {

	// TODO : fix
	//if common.GetConfigInstance().Debug {
	//	thread.logger.Println("received ping over C15")
	//}
	//
	//response := common.DatabaseResponse{
	//	Nonce:   request.Nonce,
	//	Success: true,
	//}
	//thread.C <- response
}

func (thread *Thread) ProcessIncomingResponse(response *common.MessengerResponse) {
	thread.messengerResponseTable.Write(response.Nonce, response)
}
