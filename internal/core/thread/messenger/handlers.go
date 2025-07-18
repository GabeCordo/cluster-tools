package messenger

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/component/message"
	"github.com/GabeCordo/Flock/internal/core/component/message/log"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

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
