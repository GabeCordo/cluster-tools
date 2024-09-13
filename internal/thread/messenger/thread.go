package messenger

import (
	"github.com/GabeCordo/cluster-tools/internal/message"
	"github.com/GabeCordo/cluster-tools/internal/message/log"
	"github.com/GabeCordo/cluster-tools/internal/thread"
)

func (th *Thread) Setup() {
	th.accepting = true
}

func (th *Thread) Start() {

	// LISTEN TO INCOMING REQUESTS

	go func() {
		// request coming from database
		for request := range th.C3 {
			if !th.accepting {
				break
			}
			th.wg.Add(1)

			request.Source = thread.Database
			th.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// request coming from supervisor
		for request := range th.C17 {
			if !th.accepting {
				break
			}
			th.wg.Add(1)

			request.Source = thread.Supervisor
			th.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// request coming from supervisor
		for request := range th.C22 {
			if !th.accepting {
				break
			}
			th.wg.Add(1)

			request.Source = thread.HttpClient
			th.ProcessIncomingRequest(&request)
		}
	}()
}

func (th *Thread) Respond(request *thread.Request, response *thread.Response) (success bool) {

	success = true

	switch request.Source {
	case thread.Database:
		th.C4 <- *response
	case thread.HttpClient:
		th.C23 <- *response
	default:
		success = false
	}

	return success
}

func (th *Thread) ProcessIncomingRequest(request *thread.Request) {

	response := &thread.Response{Nonce: request.Nonce, Source: thread.Messenger}

	switch request.Action {
	case thread.GetAction:
		switch request.Type {
		case thread.SmtpRecord:
			th.logger.Warn("SMTP record get called BUT is not implemented!")
		default:
			response.Error = thread.BadRequestType
		}
	case thread.CloseAction:
		th.ProcessCloseLogRequest(request)
	default:
		th.ProcessConsoleRequest(request)
	}

	th.Respond(request, response)
	th.wg.Done()
}

func (th *Thread) ProcessConsoleRequest(request *thread.Request) {
	var priority message.Priority

	switch request.Type {
	case thread.DefaultLogRecord:
		priority = message.Normal
	case thread.WarningLogRecord:
		priority = message.Warning
	default:
		priority = message.Fatal
	}

	th.messenger.Message(
		message.Source{
			Module:     request.Identifiers.Module,
			Cluster:    request.Identifiers.Cluster,
			Identifier: request.Identifiers.Supervisor,
		},
		log.Log{
			Priority: priority,
			Message:  (request.Data).(string),
		},
	)
}

func (th *Thread) ProcessCloseLogRequest(request *thread.Request) {

	th.logger.Printf("closing log for %s/%s\n", request.Identifiers.Module, request.Identifiers.Cluster)
	th.messenger.Flush(
		message.Source{
			Module:     request.Identifiers.Module,
			Cluster:    request.Identifiers.Cluster,
			Identifier: request.Identifiers.Supervisor,
		},
		nil,
	)
}

func (th *Thread) Teardown() {
	th.accepting = false

	th.wg.Wait()
}
