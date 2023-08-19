package messenger

import (
	"github.com/GabeCordo/mango-core/core/components/messenger"
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango/threads"
)

func (messengerThread *Thread) Setup() {
	messengerThread.accepting = true
}

func (messengerThread *Thread) Start() {

	// LISTEN TO INCOMING REQUESTS

	go func() {
		// request coming from database
		for request := range messengerThread.C3 {
			if !messengerThread.accepting {
				break
			}
			messengerThread.wg.Add(1)

			request.Source = threads.Database
			messengerThread.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// request coming from supervisor
		for request := range messengerThread.C17 {
			if !messengerThread.accepting {
				break
			}
			messengerThread.wg.Add(1)

			request.Source = threads.Supervisor
			messengerThread.ProcessIncomingRequest(&request)
		}
	}()

	messengerThread.wg.Wait()
}

func (messengerThread *Thread) Respond(module threads.Module, response any) (success bool) {

	success = true

	switch module {
	case threads.Database:
		messengerThread.C4 <- *(response).(*threads.MessengerResponse)
	default:
		success = false
	}

	return success
}

func (messengerThread *Thread) ProcessIncomingRequest(request *threads.MessengerRequest) {

	switch request.Action {
	case threads.MessengerClose:
		messengerThread.ProcessCloseLogRequest(request)
	case threads.MessengerUpperPing:
		messengerThread.ProcessMessengerPing(request)
	default:
		messengerThread.ProcessConsoleRequest(request)
	}

	messengerThread.wg.Done()
}

func (messengerThread *Thread) ProcessMessengerPing(request *threads.MessengerRequest) {

	if common.GetConfigInstance().Debug {
		messengerThread.logger.Println("[etl_messenger] received ping over C3")
	}

	response := &threads.MessengerResponse{Nonce: request.Nonce, Success: true}
	messengerThread.Respond(threads.Database, response)
}

func (messengerThread *Thread) ProcessConsoleRequest(request *threads.MessengerRequest) {
	messengerInstance := GetMessengerInstance()

	var priority messenger.MessagePriority

	switch request.Action {
	case threads.MessengerLog:
		priority = messenger.Log
	case threads.MessengerWarning:
		priority = messenger.Warning
	default:
		priority = messenger.Fatal
	}

	messengerInstance.Log(request.Module+"_"+request.Cluster, request.Message, priority)
}

func (messengerThread *Thread) ProcessCloseLogRequest(request *threads.MessengerRequest) {
	messengerInstance := GetMessengerInstance()
	messengerInstance.Complete(request.Module + "_" + request.Cluster)
}

func (messengerThread *Thread) Teardown() {
	messengerThread.accepting = false

	messengerThread.wg.Wait()
}
