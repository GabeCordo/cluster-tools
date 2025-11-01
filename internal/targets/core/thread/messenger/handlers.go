package messenger

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/targets/core/component/message"
	"github.com/FortifiedCode/flock/internal/targets/core/component/message/log"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
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

	data, ok := (request.Data).(string)
	if !ok {
		return errors.New("expected request.Data to be of type string")
	}

	err := th.messenger.Message(
		message.Source{
			Module:     request.Identifiers.Module,
			Cluster:    request.Identifiers.Function,
			Identifier: request.Identifiers.Supervisor,
		},
		log.Log{
			Priority: priority,
			Message:  data,
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
