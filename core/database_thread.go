package core

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/components/database"
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
		db.C2 <- *response
		break
	case Provisioner:
		db.C8 <- *response
		break
	}
}

func (db *DatabaseThread) ProcessIncomingRequest(request *DatabaseRequest) {
	d := GetDatabaseInstance()

	switch request.Action {
	case Store:
		{
			switch request.Type {
			case database.Config:
				{
					configData := (request.Data).(cluster.Config)
					isOk := d.StoreClusterConfig(configData)
					db.Send(request, &DatabaseResponse{Success: isOk})
				}
			case database.Statistic:
				{
					statisticsData := (request.Data).(*cluster.Response)
					isOk := d.StoreUsageRecord(request.Cluster, statisticsData.Stats, statisticsData.LapsedTime)
					db.Send(request, &DatabaseResponse{Success: isOk})
				}
			}
		}
	case Fetch:
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
	}

	db.wg.Done()
}

func (db *DatabaseThread) ProcessIncomingResponse(response *MessengerResponse) {

}

func (db *DatabaseThread) Teardown() {
	db.accepting = false

	db.wg.Wait()
}
