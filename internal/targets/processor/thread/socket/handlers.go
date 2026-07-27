package socket

import (
	common "github.com/GabeCordo/FunctionScheduler/internal/shared/socket"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/run"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/processor/thread"
)

func (t *Thread) handleSocketModuleAdd(request *thread.SocketRequest) {

	module, ok := request.Data.(ScalingFunctions.ModuleIR)
	if !ok {
		t.logger.Warnln("received module add with invalid data, expected ScalingFunctions.ModuleIR")
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
