package messenger

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/component/message"
	"github.com/GabeCordo/Flock/internal/core/component/message/log"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (th *Thread) Setup() {

}

func (th *Thread) Start() {

	var iReq *thread.Request
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-th.channels.c3:
			{
				oRsp = th.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					th.channels.c4 <- oRsp
				}
			}
		case iReq = <-th.channels.c22:
			{
				oRsp = th.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					th.channels.c23 <- oRsp
				}
			}
		case iReq = <-th.channels.c17:
			{
				_ = th.handleRequest(iReq)
			}
		case <-th.channels.close:
			{
				// shutting down the messenger thread
				break
			}
		}
		oRsp = nil
	}
}

func (th *Thread) handleRequest(request *thread.Request) (response *thread.Response) {

	var err error

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.SmtpRecord:
				{
					// SMTP record get called BUT is not implemented
					err = thread.NotImplemented
				}
			default:
				{
					err = thread.BadRequestType
				}
			}
		}
	case thread.CloseAction:
		{
			err = th.ProcessCloseLogRequest(request)
		}
	default:
		{
			err = th.ProcessConsoleRequest(request)
		}
	}

	response = thread.NewResponse(thread.Messenger)
	response.Error = err
	response.Success = err == nil
	return response
}

func (th *Thread) ProcessConsoleRequest(request *thread.Request) error {
	var priority message.Priority

	switch request.Type {
	case thread.DefaultLogRecord:
		priority = message.Normal
	case thread.WarningLogRecord:
		priority = message.Warning
	default:
		priority = message.Fatal
	}

	err := th.messenger.Message(
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
	return err
}

func (th *Thread) ProcessCloseLogRequest(request *thread.Request) error {

	th.logger.Printf("[%s][%s][%d] closing log\n",
		request.Identifiers.Module,
		request.Identifiers.Function,
		request.Identifiers.Supervisor,
	)
	err := th.messenger.Flush(
		message.Source{
			Module:     request.Identifiers.Module,
			Cluster:    request.Identifiers.Function,
			Identifier: request.Identifiers.Supervisor,
		},
		nil,
	)

	// Concept: An error can indicate a module, pipeline, or runner was not found in the messenger.
	//			This happens when a run (on a processor) never sends a log to the core.
	//
	// Action: Only log other types of errors encountered.
	if errors.Is(err, message.LogSaveFailedError) {
		th.logger.Printf("closing log failed %s\n", err.Error())
	}

	return err
}

func (th *Thread) Teardown() {

	th.wg.Wait()

	// send a notification to the Start() goroutine to terminate
	th.channels.close <- thread.Shutdown
}
