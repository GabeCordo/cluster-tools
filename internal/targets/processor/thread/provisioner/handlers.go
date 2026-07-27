package provisioner

import (
	"errors"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/run"
	thread2 "github.com/GabeCordo/FunctionScheduler/internal/targets/processor/thread"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/processor/use_cases/provisioner"
)

func (t *Thread) handleGetModules(request *thread2.ProvisionerRequest, response *thread2.ProvisionerResponse) {

	response.Error = errors.New("implement me")
}

func (t *Thread) handleCreateRun(request *thread2.ProvisionerRequest, response *thread2.ProvisionerResponse) {

	// Note: configs are now sent from the core, we don't need to worry about looking for, verifying, or
	//		 reverting to a default cluster.pipeline if one is not provided
	if request == nil || request.Pipeline == nil {
		response.Error = errors.New("passed value is nil")
		return
	}

	provisionRequest := provisioner.ProvisionRequest{
		Pipeline:   request.Pipeline,
		Supervisor: request.Supervisor,
		Metadata:   request.Metadata,
		Namespace:  request.Namespace,
		Core:       *t.Config.Core, // TODO: fix unsafe deref
	}

	response.Error = t.useCases.CreateRun(&provisionRequest, func(r *run.Run) {
		mandatory := thread2.SocketMandatory{Pipe: t.channels.C0, NoncePool: t.noncePool}
		thread2.AsyncRunUpdate(mandatory, r)
	})
}

func (t *Thread) handleStopRun(request *thread2.ProvisionerRequest, response *thread2.ProvisionerResponse) {

	response.Error = t.useCases.StopRun(request.Supervisor)
}

func (t *Thread) handleGetStatistics(request *thread2.ProvisionerRequest, response *thread2.ProvisionerResponse) {

	response.Data = t.useCases.GetStatistics()
}

func (t *Thread) handleRegisterModulesToCore(request *thread2.ProvisionerRequest, response *thread2.ProvisionerResponse) {

	mandatory := thread2.SocketMandatory{
		Pipe:      t.channels.C0,
		NoncePool: t.noncePool,
	}
	for _, moduleInst := range t.useCases.GetModules() {
		ir := moduleInst.GetIR()
		thread2.AsyncModuleAdd(mandatory, ir)
	}
}
