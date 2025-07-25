package socket

import (
	"encoding/json"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/thread"
	common "github.com/GabeCordo/Flock/internal/shared/async"
)

func (t *Thread) handleCreateRun(request *thread.Request, response *thread.Response) {

	if request.Identifiers.Processor == 0 {
		t.logger.Warnln("processor identifier missing for create run")
		response.Error = thread.BadRequestType
		return
	}

	conn, err := t.useCases.GetProcessor(request.Identifiers.Processor)
	if err != nil {
		t.logger.Warnln(err.Error())
		response.Error = thread.BadRequestType
		return
	}

	runRequest, ok := request.Data.(run.Request)
	if !ok {
		t.logger.Warnln("create run was not given a run.Request type")
		response.Error = thread.BadRequestType
		return
	}

	encoder := json.NewEncoder(conn)

	r := common.Request{
		Action: common.Create,
		Record: common.Run,
		Data:   runRequest,
	}

	err = encoder.Encode(r)
	if err != nil {
		t.logger.Warnln("failed to encode run request")
		response.Error = thread.InternalError
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

	r := common.Request{
		Action: common.Delete,
		Record: common.Run,
		Data:   request.Identifiers.Supervisor,
	}

	conn, err := t.useCases.GetProcessor(request.Identifiers.Processor)
	if err != nil {
		t.logger.Warnln(err.Error())
		response.Error = thread.BadRequestType
		return
	}

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(r)
	if err != nil {
		t.logger.Warnln("failed to encode delete run request")
	}
}
