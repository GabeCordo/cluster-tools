package messenger

import (
	"github.com/GabeCordo/cluster-tools/core/components/messenger"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
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

	go func() {
		// request coming from supervisor
		for request := range thread.C22 {
			if !thread.accepting {
				break
			}
			thread.wg.Add(1)

			request.Source = common.HttpClient
			thread.ProcessIncomingRequest(&request)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) Respond(request *common.ThreadRequest, response *common.ThreadResponse) (success bool) {

	success = true

	switch request.Source {
	case common.Database:
		thread.C4 <- *response
	case common.HttpClient:
		thread.C23 <- *response
	default:
		success = false
	}

	return success
}

func (thread *Thread) ProcessIncomingRequest(request *common.ThreadRequest) {

	response := &common.ThreadResponse{Nonce: request.Nonce, Source: common.Messenger}

	switch request.Action {
	case common.GetAction:
		switch request.Type {
		case common.SmtpRecord:
			response.Data = thread.getSmtpRecord()
		default:
			response.Error = common.BadRequestType
		}
	case common.CloseAction:
		thread.ProcessCloseLogRequest(request)
	default:
		thread.ProcessConsoleRequest(request)
	}

	thread.Respond(request, response)
	thread.wg.Done()
}

func (thread *Thread) ProcessConsoleRequest(request *common.ThreadRequest) {
	messengerInstance := GetMessengerInstance(thread.config)

	var priority messenger.MessagePriority

	switch request.Type {
	case common.DefaultLogRecord:
		priority = messenger.Normal
	case common.WarningLogRecord:
		priority = messenger.Warning
	default:
		priority = messenger.Fatal
	}

	messengerInstance.Log(
		request.Identifiers.Module,
		request.Identifiers.Cluster,
		request.Identifiers.Supervisor,
		(request.Data).(string),
		priority,
	)
}

func (thread *Thread) ProcessCloseLogRequest(request *common.ThreadRequest) {

	thread.logger.Printf("closing log for %s/%s\n", request.Identifiers.Module, request.Identifiers.Cluster)
	messengerInstance := GetMessengerInstance(thread.config)
	messengerInstance.Complete(request.Identifiers.Module, request.Identifiers.Cluster, request.Identifiers.Supervisor)
}

func (thread *Thread) Teardown() {
	thread.accepting = false

	thread.wg.Wait()
}
