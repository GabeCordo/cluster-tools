package core

import (
	"fmt"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/components/database"
	"math/rand"
	"time"
)

var DatabaseInstance *database.Database

func GetDatabaseInstance() *database.Database {
	if DatabaseInstance == nil {
		DatabaseInstance = database.NewDatabase()
	}

	return DatabaseInstance
}

func (databaseThread *DatabaseThread) Setup() {
	databaseThread.accepting = true
}

func (databaseThread *DatabaseThread) Start() {

	for databaseThread.accepting {

		select {
		case request := <-databaseThread.C1: // request from http_server
			request.Origin = threads.Http
			databaseThread.wg.Add(1)
			go databaseThread.ProcessIncomingRequest(&request)
		case request := <-databaseThread.C7: // request from provisioner
			request.Origin = threads.Provisioner
			databaseThread.wg.Add(1)
			go databaseThread.ProcessIncomingRequest(&request)
		case response := <-databaseThread.C4:
			go databaseThread.ProcessIncomingResponse(&response)
		default:
			time.Sleep(1 * time.Second)
		}
	}

	databaseThread.wg.Wait()
}

func (databaseThread *DatabaseThread) Send(request *threads.DatabaseRequest, response *threads.DatabaseResponse) {
	switch request.Origin {
	case threads.Http:
		databaseThread.C2 <- *response
		break
	case threads.Provisioner:
		databaseThread.C8 <- *response
		break
	}
}

func (databaseThread *DatabaseThread) ProcessIncomingRequest(request *threads.DatabaseRequest) {
	d := GetDatabaseInstance()

	switch request.Action {
	case threads.DatabaseStore:
		{
			switch request.Type {
			case threads.ClusterConfig:
				{
					configData := (request.Data).(cluster.Config)
					isOk := d.StoreClusterConfig(request.Module, configData)

					databaseThread.Send(request, &threads.DatabaseResponse{Success: isOk, Nonce: request.Nonce})
				}
			case threads.SupervisorStatistic:
				{
					statisticsData := (request.Data).(*cluster.Response)
					isOk := d.StoreUsageRecord(request.Module, request.Cluster, statisticsData.Stats, statisticsData.LapsedTime)

					databaseThread.Send(request, &threads.DatabaseResponse{Success: isOk, Nonce: request.Nonce})
				}
			}
		}
	case threads.DatabaseFetch:
		{
			var response threads.DatabaseResponse

			switch request.Type {
			case threads.ClusterConfig:
				{
					config, ok := d.GetClusterConfig(request.Module, request.Cluster)
					if !ok {
						response = threads.DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = threads.DatabaseResponse{Success: true, Nonce: request.Nonce, Data: config}
					}

					databaseThread.Send(request, &response)
				}
			case threads.SupervisorStatistic:
				{
					record, ok := d.GetUsageRecord(request.Module, request.Cluster)
					if !ok {
						response = threads.DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = threads.DatabaseResponse{Success: true, Nonce: request.Nonce, Data: record.Entries[:record.Head+1]}
					}

					databaseThread.Send(request, &response)
				}
			}
		}
	case threads.DatabaseDelete:
		{
			switch request.Type {
			case threads.ClusterModule:
				{
					success := d.DeleteModuleRecords(request.Module)

					response := threads.DatabaseResponse{Success: success, Nonce: request.Nonce}
					databaseThread.Send(request, &response)
				}
			}
		}
	case threads.DatabaseReplace:
		{
			config := (request.Data).(cluster.Config)
			success := d.ReplaceClusterConfig(request.Module, config)
			response := threads.DatabaseResponse{Success: success, Nonce: request.Nonce}

			databaseThread.Send(request, &response)
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

func (databaseThread *DatabaseThread) ProcessDatabaseUpperPing(request *threads.DatabaseRequest) {

	if GetConfigInstance().Debug {
		databaseThread.logger.Println("received ping over C1")
	}

	messengerPingRequest := threads.MessengerRequest{
		Action: threads.MessengerUpperPing,
		Nonce:  rand.Uint32(),
	}
	databaseThread.C3 <- messengerPingRequest

	messengerTimeout := false
	var messengerResponse *threads.MessengerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			messengerTimeout = true
			break
		}

		if responseEntry, found := databaseThread.messengerResponseTable.Lookup(messengerPingRequest.Nonce); found {
			messengerResponse = (responseEntry).(*threads.MessengerResponse)
			break
		}
	}

	fmt.Println(messengerTimeout)
	fmt.Println(messengerResponse)

	if messengerTimeout {
		databaseThread.C2 <- threads.DatabaseResponse{
			Nonce:   request.Nonce,
			Success: false,
		}
		return
	}

	if GetConfigInstance().Debug && messengerResponse.Success {
		databaseThread.logger.Println("received ping over C4")
	}

	databaseThread.C2 <- threads.DatabaseResponse{
		Nonce:   request.Nonce,
		Success: messengerResponse.Success,
	}
}

func (databaseThread *DatabaseThread) ProcessDatabaseLowerPing(request *threads.DatabaseRequest) {

	if GetConfigInstance().Debug {
		databaseThread.logger.Println("received ping over C7")
	}

	databaseThread.C8 <- threads.DatabaseResponse{
		Nonce:   request.Nonce,
		Success: true,
	}
	fmt.Println("done")
}

func (databaseThread *DatabaseThread) ProcessIncomingResponse(response *threads.MessengerResponse) {
	databaseThread.messengerResponseTable.Write(response.Nonce, response)
}

func (databaseThread *DatabaseThread) Teardown() {
	databaseThread.accepting = false

	databaseThread.wg.Wait()
}
