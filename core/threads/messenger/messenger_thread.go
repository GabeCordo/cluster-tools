package messenger

import (
	"github.com/GabeCordo/mango/core/components/messenger"
	"github.com/GabeCordo/mango/core/threads/common"
)

func (thread *Thread) Setup() {
	thread.accepting = true
}

func (thread *Thread) Start() {

	// LISTEN TO INCOMING REQUESTS

	go func() {
		// request coming from database
		for request := range thread.C3 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.Database
			thread.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// request coming from supervisor
		for request := range thread.C17 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.Supervisor
			thread.ProcessIncomingRequest(&request)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) Respond(module common.Module, response any) (success bool) {

	success = true

	switch module {
	case common.Database:
		thread.C4 <- *(response).(*common.MessengerResponse)
	default:
		success = false
	}

	return success
}

func (thread *Thread) ProcessIncomingRequest(request *common.MessengerRequest) {

	switch request.Action {
	case common.MessengerClose:
		thread.ProcessCloseLogRequest(request)
	case common.MessengerUpperPing:
		thread.ProcessMessengerPing(request)
	default:
		thread.ProcessConsoleRequest(request)
	}

	thread.wg.Done()
}

func (thread *Thread) ProcessMessengerPing(request *common.MessengerRequest) {

	if thread.config.Debug {
		thread.logger.Println("[etl_messenger] received ping over C3")
	}

	response := &common.MessengerResponse{Nonce: request.Nonce, Success: true}
	thread.Respond(common.Database, response)
}

func (thread *Thread) ProcessConsoleRequest(request *common.MessengerRequest) {
	messengerInstance := GetMessengerInstance(thread.config)

	var priority messenger.MessagePriority

	switch request.Action {
	case common.MessengerLog:
		priority = messenger.Normal
	case common.MessengerWarning:
		priority = messenger.Warning
	default:
		priority = messenger.Fatal
	}

	messengerInstance.Log(request.Module, request.Cluster, request.Supervisor, request.Message, priority)
}

func (thread *Thread) ProcessCloseLogRequest(request *common.MessengerRequest) {

	thread.logger.Printf("closing log for %s/%s\n", request.Module, request.Cluster)
	messengerInstance := GetMessengerInstance(thread.config)
	messengerInstance.Complete(request.Module, request.Cluster, request.Supervisor)
}

func (thread *Thread) Teardown() {
	thread.accepting = false

	thread.wg.Wait()
}
