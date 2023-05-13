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

func (db *DatabaseThread) Setup() {
	db.accepting = true
}

func (db *DatabaseThread) Start() {
	go func() {
		// request from http_server
		for request := range db.C1 {
			if !db.accepting {
				break
			}
			db.wg.Add(1)
			db.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		// request from supervisor
		for request := range db.C7 {
			if !db.accepting {
				break
			}
			db.wg.Add(1)
			db.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		for response := range db.C4 {
			if !db.accepting {
				break
			}
			db.ProcessIncomingResponse(&response)
		}
	}()

	db.wg.Wait()
}

func (db *DatabaseThread) Send(request *DatabaseRequest, response *DatabaseResponse) {
	switch request.Origin {
	case Http:
		fmt.Println("sent response to Http")
		db.C2 <- *response
		break
	case Provisioner:
		fmt.Println("sent response to Provisioner")
		db.C8 <- *response
		break
	}
}

func (db *DatabaseThread) ProcessIncomingRequest(request *DatabaseRequest) {
	d := GetDatabaseInstance()

	switch request.Action {
	case DatabaseStore:
		{
			switch request.Type {
			case database.Config:
				{
					configData := (request.Data).(cluster.Config)
					isOk := d.StoreClusterConfig(configData)
					db.Send(request, &DatabaseResponse{Success: isOk, Nonce: request.Nonce})
				}
			case database.Statistic:
				{
					statisticsData := (request.Data).(*cluster.Response)
					isOk := d.StoreUsageRecord(request.Cluster, statisticsData.Stats, statisticsData.LapsedTime)
					db.Send(request, &DatabaseResponse{Success: isOk, Nonce: request.Nonce})
				}
			}
		}
	case DatabaseFetch:
		{
			var response DatabaseResponse

			switch request.Type {
			case database.Config:
				{
					fmt.Println("getting config")

					config, ok := d.GetClusterConfig(request.Cluster)
					if !ok {
						response = DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = DatabaseResponse{Success: true, Nonce: request.Nonce, Data: config}
					}

					db.Send(request, &response)
				}
			case database.Statistic:
				{
					record, ok := d.GetUsageRecord(request.Cluster)
					if !ok {
						response = DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = DatabaseResponse{Success: true, Nonce: request.Nonce, Data: record.Entries[:record.Head+1]}
					}

					db.Send(request, &response)
				}
			}
		}
	case DatabaseUpperPing:
		{
			db.ProcessDatabaseUpperPing(request)
		}
	case DatabaseLowerPing:
		{
			db.ProcessDatabaseLowerPing(request)
		}
	}

	db.wg.Done()
}

func (db *DatabaseThread) ProcessDatabaseUpperPing(request *DatabaseRequest) {

	if GetConfigInstance().Debug {
		log.Println("[etl_database] received ping over C1")
	}

	messengerPingRequest := MessengerRequest{Action: MessengerUpperPing, Nonce: rand.Uint32()}
	db.C3 <- messengerPingRequest

	messengerTimeout := false
	var messengerResponse *MessengerResponse

	timestamp := time.Now()
	for {
		if time.Now().Sub(timestamp).Seconds() > 2.0 {
			messengerTimeout = true
			break
		}

		if responseEntry, found := db.messengerResponseTable.Lookup(messengerPingRequest.Nonce); found {
			messengerResponse = (responseEntry).(*MessengerResponse)
			break
		}
	}

	if GetConfigInstance().Debug && (!messengerTimeout || messengerResponse.Success) {
		log.Println("[etl_database] received ping over C4")
	}

	db.C2 <- DatabaseResponse{Nonce: request.Nonce, Success: messengerTimeout || messengerResponse.Success}
}

func (db *DatabaseThread) ProcessDatabaseLowerPing(request *DatabaseRequest) {

	if GetConfigInstance().Debug {
		log.Println("[etl_database] received ping over C7")
	}

	db.C8 <- DatabaseResponse{Nonce: request.Nonce, Success: true}
}

func (db *DatabaseThread) ProcessIncomingResponse(response *MessengerResponse) {
	db.messengerResponseTable.Write(response.Nonce, response)
}

func (db *DatabaseThread) Teardown() {
	db.accepting = false

	db.wg.Wait()
}
