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
	db.wg.Add(1)

	go func() {
		for db.accepting {
			request := <-db.C1 // request from http_server
			db.ProcessIncomingRequest(&request)
		}
		db.wg.Done() // only the HTTP server can call an interrupt
	}()
	go func() {
		for db.accepting {
			request := <-db.C7 // request from supervisor
			db.ProcessIncomingRequest(&request)
		}
	}()
	go func() {
		for db.accepting {
			response := <-db.C4 // responses from messenger
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
	switch request.Action {
	case Store:
		{
			d := GetDatabaseInstance()
			d.Store(request.Cluster, request.Data)

			response := DatabaseResponse{Success: true}
			db.Send(request, &response)
		}
	case Peak:
		{
			d := GetDatabaseInstance()
			record := d.PeakRecord(request.Cluster)

			var response DatabaseResponse
			if record.Head == -1 {
				response = DatabaseResponse{Success: false, Nonce: request.Nonce}
			} else {
				response = DatabaseResponse{Success: true, Nonce: request.Nonce, Data: record.Entries[:record.Head]}
			}

			db.Send(request, &response)
		}
	case Fetch:
		{
			d := GetDatabaseInstance()
			record := d.GetRecord(request.Cluster)

			var response DatabaseResponse
			if record.Head == -1 {
				response = DatabaseResponse{Success: false, Nonce: request.Nonce}
			} else {
				response = DatabaseResponse{Success: true, Nonce: request.Nonce, Data: record.Entries[:record.Head+1]}
			}

			db.Send(request, &response)
		}
	}
}

func (db *DatabaseThread) ProcessIncomingResponse(response *MessengerResponse) {

}

func (db *DatabaseThread) Teardown() {
	db.accepting = false
}
