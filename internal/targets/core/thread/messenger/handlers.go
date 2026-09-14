package messenger

import (
	"errors"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/message"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/message/log"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/thread"
)

func (t *Thread) ProcessConsoleRequest(request *thread.Request) error {
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

	err := t.messenger.Message(
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

func (t *Thread) ProcessCloseLogRequest(request *thread.Request) error {

	t.logger.Printf("[%s][%s][%d] closing log\n",
		request.Identifiers.Module,
		request.Identifiers.Function,
		request.Identifiers.Supervisor,
	)
	err := t.messenger.Flush(
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
		t.logger.Printf("closing log failed %s\n", err.Error())
	}

	return err
}
