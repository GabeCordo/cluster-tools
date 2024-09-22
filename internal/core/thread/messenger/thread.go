package messenger

import (
	"github.com/GabeCordo/cluster-tools/internal/core/message"
	"github.com/GabeCordo/cluster-tools/internal/core/message/log"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
)

func (th *Thread) Setup() {
	th.accepting = true
}

func (th *Thread) Start() {

	// LISTEN TO INCOMING REQUESTS

	thread.SetupListener(th.C3, th.C4, &th.accepting, &th.wg, thread.Messenger, th.Handle)

	thread.SetupListener(th.C17, nil, &th.accepting, &th.wg, thread.Messenger, th.Handle)

	thread.SetupListener(th.C22, th.C23, &th.accepting, &th.wg, thread.Messenger, th.Handle)
}

func (th *Thread) Handle(request *thread.Request, response *thread.Response) {

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.SmtpRecord:
				{
					th.logger.Warn("SMTP record get called BUT is not implemented!")
				}
			default:
				{
					response.Error = thread.BadRequestType
				}
			}
		}
	case thread.CloseAction:
		{
			th.ProcessCloseLogRequest(request)
		}
	default:
		{
			th.ProcessConsoleRequest(request)
		}
	}
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
			Cluster:    request.Identifiers.Function,
			Identifier: request.Identifiers.Supervisor,
		},
		log.Log{
			Priority: priority,
			Message:  (request.Data).(string),
		},
	)
}

func (th *Thread) ProcessCloseLogRequest(request *thread.Request) {

	th.logger.Printf("closing log for %s/%s\n", request.Identifiers.Module, request.Identifiers.Function)
	th.messenger.Flush(
		message.Source{
			Module:     request.Identifiers.Module,
			Cluster:    request.Identifiers.Function,
			Identifier: request.Identifiers.Supervisor,
		},
		nil,
	)
}

func (th *Thread) Teardown() {
	th.accepting = false

	th.wg.Wait()
}
