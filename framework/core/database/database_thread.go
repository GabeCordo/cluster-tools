package database

import (
	"fmt"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/components/database"
	"github.com/GabeCordo/etl/framework/core/common"
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

func (databaseThread *Thread) Setup() {
	databaseThread.accepting = true
}

func (databaseThread *Thread) Start() {

	go func() {
		// request from http_server
		for request := range databaseThread.C1 {
			if !databaseThread.accepting {
				break
			}
			request.Origin = threads.Http
			databaseThread.wg.Add(1)
			databaseThread.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range databaseThread.C7 {
			if !databaseThread.accepting {
				break
			}
			request.Origin = threads.Provisioner
			databaseThread.wg.Add(1)
			databaseThread.ProcessIncomingRequest(&request)
		}
	}()
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

func (databaseThread *Thread) Send(request *threads.DatabaseRequest, response *threads.DatabaseResponse) {
	switch request.Origin {
	case threads.Http:
		databaseThread.C2 <- *response
		break
	case threads.Provisioner:
		databaseThread.C8 <- *response
		break
	}
}

func (databaseThread *Thread) ProcessIncomingRequest(request *threads.DatabaseRequest) {
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

func (databaseThread *Thread) ProcessDatabaseUpperPing(request *threads.DatabaseRequest) {

	fmt.Printf("got from http (%d)\n", request.Nonce)

	if common.GetConfigInstance().Debug {
		databaseThread.logger.Println("received ping over C1")
	}

	messengerPingRequest := threads.MessengerRequest{
		Action: threads.MessengerUpperPing,
		Nonce:  rand.Uint32(),
	}
	fmt.Printf("send to msg (%d)\n", messengerPingRequest.Nonce)
	databaseThread.C3 <- messengerPingRequest

	messengerTimeout := false
	var messengerResponse *threads.MessengerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > common.GetConfigInstance().MaxWaitForResponse {
			messengerTimeout = true
			break
		}

		if responseEntry, found := databaseThread.messengerResponseTable.Lookup(messengerPingRequest.Nonce); found {
			messengerResponse = (responseEntry).(*threads.MessengerResponse)
			break
		}
	}

	if messengerTimeout {
		databaseThread.C2 <- threads.DatabaseResponse{
			Nonce:   request.Nonce,
			Success: false,
		}
		return
	}

	fmt.Printf("got from msg (%d)(%t)\n", messengerResponse.Nonce, messengerResponse.Success)
	if common.GetConfigInstance().Debug && messengerResponse.Success {
		databaseThread.logger.Println("received ping over C4")
	}

	databaseResponse := threads.DatabaseResponse{
		Nonce:   request.Nonce,
		Success: messengerResponse.Success,
	}
	fmt.Printf("send to http (%d)\n", databaseResponse.Nonce)
	databaseThread.C2 <- databaseResponse
}

func (databaseThread *Thread) ProcessDatabaseLowerPing(request *threads.DatabaseRequest) {

	fmt.Printf("got from prov (%d)\n", request.Nonce)
	if common.GetConfigInstance().Debug {
		databaseThread.logger.Println("received ping over C7")
	}

	response := threads.DatabaseResponse{
		Nonce:   request.Nonce,
		Success: true,
	}
	fmt.Printf("send to prov (%d, %t)\n", response.Nonce, response.Success)
	databaseThread.C8 <- response
}

func (databaseThread *Thread) ProcessIncomingResponse(response *threads.MessengerResponse) {
	databaseThread.messengerResponseTable.Write(response.Nonce, response)
}

func (databaseThread *Thread) Teardown() {
	databaseThread.accepting = false

	databaseThread.wg.Wait()
}
