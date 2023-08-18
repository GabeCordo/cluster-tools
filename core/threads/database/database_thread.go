package database

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl-light/utils"
	"github.com/GabeCordo/etl/core/components/database"
	"github.com/GabeCordo/etl/core/threads/common"
	"log"
	"math/rand"
	"time"
)

func (databaseThread *Thread) Setup() {
	databaseThread.accepting = true

	if err := GetConfigDatabaseInstance().Load(databaseThread.configFolderPath); err != nil {
		log.Panicf("could not load saved configs, run 'etl doctor' to verify the configuration is valid %s\n",
			err.Error())
	}
}

func (databaseThread *Thread) Teardown() {
	databaseThread.accepting = false

	if err := GetConfigDatabaseInstance().Save(databaseThread.configFolderPath); err != nil {
		log.Printf("failed to save configs created during runtime %s\n", err.Error())
	}

	if err := GetStatisticDatabaseInstance().Save(databaseThread.statisticFolderPath); err != nil {
		log.Printf("failed to save statistics created during runtime %s\n", err.Error())
	}

	databaseThread.wg.Wait()
}

func (databaseThread *Thread) Start() {

	// LISTEN FOR INCOMING REQUESTS

	go func() {
		// request from http_server
		for request := range databaseThread.C1 {
			if !databaseThread.accepting {
				break
			}
			request.Origin = threads.HttpClient
			databaseThread.wg.Add(1)
			databaseThread.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range databaseThread.C11 {
			if !databaseThread.accepting {
				break
			}
			request.Origin = threads.HttpProcessor
			databaseThread.wg.Add(1)
			databaseThread.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range databaseThread.C15 {
			if !databaseThread.accepting {
				break
			}
			request.Origin = threads.Supervisor
			databaseThread.wg.Add(1)
			databaseThread.ProcessIncomingRequest(&request)
		}
	}()

	// LISTEN FOR INCOMING RESPONSES

	go func() {
		for response := range databaseThread.C4 {
			if !databaseThread.accepting {
				break
			}
			databaseThread.ProcessIncomingResponse(&response)
		}
	}()

	databaseThread.wg.Wait()
}

func (databaseThread *Thread) Request(module threads.Module, request any) (success bool) {

	success = true

	switch module {
	case threads.Messenger:
		databaseThread.C3 <- *(request).(*threads.MessengerRequest)
	default:
		success = false
	}
	return success
}

func (databaseThread *Thread) Respond(request *threads.DatabaseRequest, response *threads.DatabaseResponse) (success bool) {

	success = true

	switch request.Origin {
	case threads.HttpClient:
		databaseThread.C2 <- *response
		break
	case threads.HttpProcessor:
		databaseThread.C12 <- *response
		break
	case threads.Supervisor:
		databaseThread.C16 <- *response
	default:
		success = false
	}

	return success
}

func (databaseThread *Thread) ProcessIncomingRequest(request *threads.DatabaseRequest) {

	switch request.Action {
	case threads.DatabaseStore:
		{
			switch request.Type {
			case threads.ClusterConfig:
				{
					configData := (request.Data).(cluster.Config)
					err := GetConfigDatabaseInstance().Create(request.Module, request.Cluster, configData)

					databaseThread.Respond(request, &threads.DatabaseResponse{
						Success: err == nil,
						Nonce:   request.Nonce,
					})
				}
			case threads.SupervisorStatistic:
				{
					statisticsData := (request.Data).(*cluster.Response)
					err := GetStatisticDatabaseInstance().Create(
						request.Module, request.Cluster,
						database.Statistic{
							Timestamp: time.Now(),
							Elapsed:   statisticsData.LapsedTime,
							Stats:     *statisticsData.Stats,
						})
					databaseThread.Respond(request, &threads.DatabaseResponse{
						Success: err == nil,
						Nonce:   request.Nonce,
					})
				}
			}
		}
	case threads.DatabaseFetch:
		{
			var response threads.DatabaseResponse

			switch request.Type {
			case threads.ClusterConfig:
				{
					config, err := GetConfigDatabaseInstance().Get(database.ConfigFilter{
						Module:     request.Module,
						Identifier: request.Cluster,
					})
					if err != nil {
						response = threads.DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						// TODO - use to expect one record, now will have many
						response = threads.DatabaseResponse{Success: true, Nonce: request.Nonce, Data: config}
					}

					databaseThread.Respond(request, &response)
				}
			case threads.SupervisorStatistic:
				{
					records, err := GetStatisticDatabaseInstance().Get(database.StatisticFilter{
						Module:  request.Module,
						Cluster: request.Cluster,
					})
					response = threads.DatabaseResponse{Success: err == nil, Nonce: request.Nonce, Data: records}
					databaseThread.Respond(request, &response)
				}
			}
		}
	case threads.DatabaseDelete:
		{
			switch request.Type {
			case threads.ClusterConfig:
				{
					err := GetConfigDatabaseInstance().Delete(request.Module, request.Cluster)
					response := threads.DatabaseResponse{Success: err == nil, Nonce: request.Nonce}
					databaseThread.Respond(request, &response)
				}
			case threads.ClusterModule:
				{
					err := GetStatisticDatabaseInstance().Delete(request.Module)

					response := threads.DatabaseResponse{Success: err == nil, Nonce: request.Nonce}
					databaseThread.Respond(request, &response)
				}
			}
		}
	case threads.DatabaseReplace:
		{
			config := (request.Data).(cluster.Config)
			err := GetConfigDatabaseInstance().Replace(request.Module, request.Cluster, config)
			response := threads.DatabaseResponse{Success: err == nil, Nonce: request.Nonce}

			databaseThread.Respond(request, &response)
		}
	case threads.DatabaseUpperPing:
		{
			databaseThread.ProcessDatabaseUpperPing(request)
		}
	case threads.DatabaseLowerPing:
		{
			databaseThread.ProcessDatabaseLowerPing(request)
		}
	}

	databaseThread.wg.Done()
}

func (databaseThread *Thread) ProcessDatabaseUpperPing(request *threads.DatabaseRequest) {

	if common.GetConfigInstance().Debug {
		databaseThread.logger.Println("received ping over C1")
	}

	messengerPingRequest := &threads.MessengerRequest{
		Action: threads.MessengerUpperPing,
		Nonce:  rand.Uint32(),
	}
	databaseThread.Request(threads.Messenger, messengerPingRequest)

	data, didTimeout := utils.SendAndWait(databaseThread.messengerResponseTable, messengerPingRequest.Nonce,
		common.GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		databaseThread.C2 <- threads.DatabaseResponse{
			Nonce:   request.Nonce,
			Success: false,
		}
		return
	}

	messengerResponse := (data).(*threads.MessengerResponse)

	if !messengerResponse.Success {
		databaseThread.C2 <- threads.DatabaseResponse{
			Nonce:   request.Nonce,
			Success: false,
		}
		return
	}

	if common.GetConfigInstance().Debug {
		databaseThread.logger.Println("received ping over C4")
	}

	databaseResponse := threads.DatabaseResponse{
		Nonce:   request.Nonce,
		Success: messengerResponse.Success,
	}
	databaseThread.C2 <- databaseResponse
}

func (databaseThread *Thread) ProcessDatabaseLowerPing(request *threads.DatabaseRequest) {

	// TODO : fix
	//if common.GetConfigInstance().Debug {
	//	databaseThread.logger.Println("received ping over C15")
	//}
	//
	//response := threads.DatabaseResponse{
	//	Nonce:   request.Nonce,
	//	Success: true,
	//}
	//databaseThread.C <- response
}

func (databaseThread *Thread) ProcessIncomingResponse(response *threads.MessengerResponse) {
	databaseThread.messengerResponseTable.Write(response.Nonce, response)
}
