package core

import (
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
	"log"
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
	go func() {
		// request from http_server
		for request := range databaseThread.C1 {
			if !databaseThread.accepting {
				break
			}
			request.Origin = Http
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
			request.Origin = Provisioner
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

func (databaseThread *DatabaseThread) Send(request *DatabaseRequest, response *DatabaseResponse) {
	switch request.Origin {
	case Http:
		databaseThread.C2 <- *response
		break
	case Provisioner:
		databaseThread.C8 <- *response
		break
	}
}

func (databaseThread *DatabaseThread) ProcessIncomingRequest(request *DatabaseRequest) {
	d := GetDatabaseInstance()

	switch request.Action {
	case DatabaseStore:
		{
			switch request.Type {
			case database.Config:
				{
					fmt.Println("storing config")
					configData := (request.Data).(cluster.Config)
					isOk := d.StoreClusterConfig(configData)

					databaseThread.Send(request, &DatabaseResponse{Success: isOk, Nonce: request.Nonce})
				}
			case database.Statistic:
				{
					statisticsData := (request.Data).(*cluster.Response)
					isOk := d.StoreUsageRecord(request.Cluster, statisticsData.Stats, statisticsData.LapsedTime)

					databaseThread.Send(request, &DatabaseResponse{Success: isOk, Nonce: request.Nonce})
				}
			}
		}
	case DatabaseFetch:
		{
			var response DatabaseResponse

			switch request.Type {
			case database.Config:
				{
					config, ok := d.GetClusterConfig(request.Cluster)
					if !ok {
						response = DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = DatabaseResponse{Success: true, Nonce: request.Nonce, Data: config}
					}

					databaseThread.Send(request, &response)
				}
			case database.Statistic:
				{
					record, ok := d.GetUsageRecord(request.Cluster)
					if !ok {
						response = DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = DatabaseResponse{Success: true, Nonce: request.Nonce, Data: record.Entries[:record.Head+1]}
					}

					databaseThread.Send(request, &response)
				}
			}
		}
	case DatabaseReplace:
		{
			config := (request.Data).(cluster.Config)
			success := d.ReplaceClusterConfig(config)
			response := DatabaseResponse{Success: success, Nonce: request.Nonce}

			databaseThread.Send(request, &response)
		}
	case DatabaseUpperPing:
		{
			databaseThread.ProcessDatabaseUpperPing(request)
		}
	case DatabaseLowerPing:
		{
			databaseThread.ProcessDatabaseLowerPing(request)
		}
	}

	databaseThread.wg.Done()
}

func (databaseThread *DatabaseThread) ProcessDatabaseUpperPing(request *DatabaseRequest) {

	if GetConfigInstance().Debug {
		log.Println("[etl_database] received ping over C1")
	}

	messengerPingRequest := MessengerRequest{Action: MessengerUpperPing, Nonce: rand.Uint32()}
	databaseThread.C3 <- messengerPingRequest

	messengerTimeout := false
	var messengerResponse *MessengerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > GetConfigInstance().MaxWaitForResponse {
			messengerTimeout = true
			break
		}

		if responseEntry, found := databaseThread.messengerResponseTable.Lookup(messengerPingRequest.Nonce); found {
			messengerResponse = (responseEntry).(*MessengerResponse)
			break
		}
	}

	if GetConfigInstance().Debug && (!messengerTimeout || messengerResponse.Success) {
		log.Println("[etl_database] received ping over C4")
	}

	databaseThread.C2 <- DatabaseResponse{Nonce: request.Nonce, Success: messengerTimeout || messengerResponse.Success}
}

func (databaseThread *DatabaseThread) ProcessDatabaseLowerPing(request *DatabaseRequest) {

	if GetConfigInstance().Debug {
		log.Println("[etl_database] received ping over C7")
	}

	databaseThread.C8 <- DatabaseResponse{Nonce: request.Nonce, Success: true}
}

func (databaseThread *DatabaseThread) ProcessIncomingResponse(response *MessengerResponse) {
	databaseThread.messengerResponseTable.Write(response.Nonce, response)
}

func (databaseThread *DatabaseThread) Teardown() {
	databaseThread.accepting = false

	databaseThread.wg.Wait()
}
