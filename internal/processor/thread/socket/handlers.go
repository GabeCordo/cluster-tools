package socket

import (
	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	common "github.com/GabeCordo/Flock/internal/shared/socket"
)

func (t *Thread) handleSocketModuleAdd(request *thread.SocketRequest) {

	module, ok := request.Data.(processor.ModuleConfig)
	if !ok {
		t.logger.Warnln("received module add with invalid data")
		return
	}

	req := common.Message{
		Action: common.Create,
		Record: common.Module,
		Data:   module,
	}

	t.useCases.SendToCore(&req)
}

func (t *Thread) handleSocketRunUpdate(request *thread.SocketRequest) {

	r, ok := request.Data.(*run.Run)
	if !ok {
		t.logger.Warnln("received run update with invalid data")
	}

	req := common.Message{
		Action: common.Update,
		Record: common.Run,
		Data:   r,
	}

	t.useCases.SendToCore(&req)
}
