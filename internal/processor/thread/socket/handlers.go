package socket

import (
	"encoding/json"
	"fmt"
	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	common "github.com/GabeCordo/Flock/internal/shared/async"
)

func (t *Thread) handleSocketModuleAdd(request *thread.SocketRequest) {

	module, ok := request.Data.(processor.ModuleConfig)
	if !ok {
		t.logger.Warnln("received module add with invalid data")
		return
	}

	req := &common.Request{
		Action: common.Create,
		Record: common.Module,
		Data:   module,
	}

	encoder := json.NewEncoder(t.connection)
	err := encoder.Encode(req)
	if err != nil {
		fmt.Println(err)
		t.logger.Warnln("failed to add module over socket")
	}
}

func (t *Thread) handleSocketRunUpdate(request *thread.SocketRequest) {

	r, ok := request.Data.(*run.Run)
	if !ok {
		t.logger.Warnln("received run update with invalid data")
	}

	req := &common.Request{
		Action: common.Update,
		Record: common.Run,
		Data:   r,
	}

	encoder := json.NewEncoder(t.connection)
	err := encoder.Encode(req) // todo : fix
	if err != nil {
		fmt.Println(err)
		t.logger.Warnln("failed to update run over socket")
	}
}

func (t *Thread) handleSocketRunCreate(request *common.Request) {

	b, err := json.Marshal(request.Data)
	if err != nil {
		t.logger.Warnln("failed to marshal the received data")
		return
	}

	runRequest := new(run.Request)
	err = json.Unmarshal(b, runRequest)
	if err != nil {
		t.logger.Warnln("received invalid data for update run")
		return
	}

	mandatory := thread.ProvisionerMandatory{
		Pipe:      t.channels.C1,
		NoncePool: t.noncePool,
	}
	thread.AsyncRunStart(mandatory, runRequest.Namespace,
		runRequest.Id, runRequest.Config, runRequest.Metadata)
}

func (t *Thread) handleSocketRunDelete(request *common.Request) {

	id, ok := request.Data.(float64)
	if !ok {
		t.logger.Warnln("run delete received value other than uint64")
		return
	}

	mandatory := thread.ProvisionerMandatory{
		Pipe:      t.channels.C1,
		NoncePool: t.noncePool,
	}
	thread.AsyncRunStop(mandatory, uint64(id))
}
