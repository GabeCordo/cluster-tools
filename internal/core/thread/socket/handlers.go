package socket

import (
	"encoding/json"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/thread"
	common "github.com/GabeCordo/Flock/internal/shared/async"
	"net"
)

func (t *Thread) handleCreateRun(request *thread.Request, response *thread.Response) {

	if request.Identifiers.Processor == 0 {
		t.Logger.Warnln("processor identifier missing for create run")
		response.Error = thread.BadRequestType
		return
	}

	// TODO: any better way to clean this up + stop using strings for lookup
	t.mutex.RLock()
	t.connectionsMux.RLock()

	var connection net.Conn
	if c, found := t.connections[request.Identifiers.Processor]; !found {
		t.Logger.Warnf("no processor exists with the identifier %d\n", request.Identifiers.Processor)
		t.mutex.RUnlock()
		response.Error = thread.BadRequestType
		t.mutex.RUnlock()
		t.connectionsMux.RUnlock()
		return
	} else {
		connection = c
		t.mutex.RUnlock()
		t.connectionsMux.RUnlock()
	}

	runRequest, ok := request.Data.(run.Request)
	if !ok {
		t.Logger.Warnln("create run was not given a run.Request type")
		response.Error = thread.BadRequestType
		return
	}

	encoder := json.NewEncoder(connection)

	r := common.Request{
		Action: common.Create,
		Record: common.Run,
		Data:   runRequest,
	}
	err := encoder.Encode(r)
	if err != nil {
		t.Logger.Warnln("failed to encode run request")
		response.Error = thread.InternalError
	}
	response.Data = request.Identifiers.Supervisor
}

func (t *Thread) handleDeleteRun(request *thread.Request, response *thread.Response) {

	if request.Identifiers.Processor == 0 {
		t.Logger.Warnln("missing processor identifier for delete run")
		response.Error = thread.InternalError
	}

	if request.Identifiers.Supervisor == 0 {
		t.Logger.Warnln("missing run id for delete run")
		response.Error = thread.InternalError
	}

	r := common.Request{
		Action: common.Delete,
		Record: common.Run,
		Data:   request.Identifiers.Supervisor,
	}

	t.connectionsMux.RLock()

	var connection net.Conn
	if c, found := t.connections[request.Identifiers.Processor]; found {
		connection = c
		t.connectionsMux.RUnlock()
	} else {
		t.Logger.Warnln("processor identifier not found for delete run")
		response.Error = thread.InternalError
		t.connectionsMux.RUnlock()
		return
	}

	encoder := json.NewEncoder(connection)
	err := encoder.Encode(r)
	if err != nil {
		t.Logger.Warnln("failed to encode delete run request")
	}
}
