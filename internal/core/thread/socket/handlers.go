package socket

import (
	"github.com/FortifiedCode/flock/internal/core/database/run"
	"github.com/FortifiedCode/flock/internal/core/thread"
	common "github.com/FortifiedCode/flock/internal/shared/socket"
)

func (t *Thread) handleCreateRun(request *thread.Request, response *thread.Response) {

	if request.Identifiers.Processor == 0 {
		t.logger.Warnln("processor identifier missing for create run")
		response.Error = thread.BadRequestType
		return
	}

	runRequest, ok := request.Data.(run.Request)
	if !ok {
		t.logger.Warnln("create run was not given a run.Message type")
		response.Error = thread.BadRequestType
		return
	}

	message := &common.Message{
		Action: common.Create,
		Record: common.Run,
		Data:   runRequest,
	}

	err := t.useCases.SendDataOnSocket(common.ConnectionId(request.Identifiers.Processor), message)
	if err != nil {
		t.logger.Warnln("failed to create run with the provided processor id")
		response.Error = thread.BadRequestType
		return
	}

	response.Data = request.Identifiers.Supervisor
}

func (t *Thread) handleDeleteRun(request *thread.Request, response *thread.Response) {

	if request.Identifiers.Processor == 0 {
		t.logger.Warnln("missing processor identifier for delete run")
		response.Error = thread.InternalError
		return
	}

	if request.Identifiers.Supervisor == 0 {
		t.logger.Warnln("missing run id for delete run")
		response.Error = thread.InternalError
		return
	}

	message := &common.Message{
		Action: common.Delete,
		Record: common.Run,
		Data:   request.Identifiers.Supervisor,
	}

	err := t.useCases.SendDataOnSocket(common.ConnectionId(request.Identifiers.Processor), message)
	if err != nil {
		t.logger.Warnln(err.Error())
		response.Error = thread.BadRequestType
	}
}
