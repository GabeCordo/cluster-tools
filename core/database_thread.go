package core

import (
	"ETLFramework/database"
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
	case Supervisor:
		db.C8 <- *response
		break
	}
}

func (db *DatabaseThread) ProcessIncomingRequest(request *DatabaseRequest) {
	d := GetDatabaseInstance()

	switch request.Action {
	case Create:
		{
			switch request.Type {
			case database.Monitors:
				{
					monitor := d.CreateMonitorRecord(request.Cluster)
					if monitor != nil {
						db.Send(request, &DatabaseResponse{Success: false})
						db.wg.Done()
						return
					}
					db.Send(request, &DatabaseResponse{Success: true})
				}
			}
		}
	case Store:
		{
			switch request.Type {
			case database.Statistics:
				{
					ok := d.StoreStatistic(request.Cluster, request.Data, request.ElapsedTime)
					if !ok {
						db.Send(request, &DatabaseResponse{Success: false})
						db.wg.Done()
						return
					}
					db.Send(request, &DatabaseResponse{Success: true})
				}
			}
		}
	case Fetch:
		{
			switch request.Type {
			case database.Statistics:
				{
					var response DatabaseResponse

					record, ok := d.GetStatisticRecord(request.Cluster)
					if !ok {
						response = DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = DatabaseResponse{Success: true, Nonce: request.Nonce, Statistics: record.Entries[:record.Head+1]}
					}

					db.Send(request, &response)
				}
			case database.Monitors:
				{
					var response DatabaseResponse

					record, ok := d.GetMonitorRecord(request.Cluster)
					if !ok {
						response = DatabaseResponse{Success: false, Nonce: request.Nonce}
					} else {
						response = DatabaseResponse{Success: true, Nonce: request.Nonce, Monitors: record.Entries[:record.Head+1]}
					}

					db.Send(request, &response)
				}
			}
		}
	case Modify:
		{
			switch request.Type {
			case database.Monitors:
				{
					pair := database.MonitorId{Identifier: request.Cluster, Id: request.Entry}
					monitor := database.Monitor{
						NumLoadRoutines:      request.Data.NumProvisionedLoadRoutines,
						NumExtractRoutines:   request.Data.NumProvisionedExtractRoutines,
						NumTransformRoutines: request.Data.NumProvisionedTransformRoutes,
					}

					ok := d.ModifyMonitor(pair, &monitor)
					if !ok {
						db.Send(request, &DatabaseResponse{Success: false})
						db.wg.Done()
						return
					}

					db.Send(request, &DatabaseResponse{Success: true})
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
