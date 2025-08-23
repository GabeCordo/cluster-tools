package provisioner

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/core/database/run"
	"github.com/FortifiedCode/flock/internal/processor/thread"
	"github.com/FortifiedCode/flock/internal/processor/use_cases/provisioner"
)

func (t *Thread) handleGetModules(request *thread.ProvisionerRequest, response *thread.ProvisionerResponse) {

	response.Error = errors.New("implement me")
}

func (t *Thread) handleCreateRun(request *thread.ProvisionerRequest, response *thread.ProvisionerResponse) {

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
		mandatory := thread.SocketMandatory{Pipe: t.channels.C0, NoncePool: t.noncePool}
		thread.AsyncRunUpdate(mandatory, r)
	})
}

func (t *Thread) handleStopRun(request *thread.ProvisionerRequest, response *thread.ProvisionerResponse) {

	response.Error = t.useCases.StopRun(request.Supervisor)
}

func (t *Thread) handleGetStatistics(request *thread.ProvisionerRequest, response *thread.ProvisionerResponse) {

	response.Data = t.useCases.GetStatistics()
}

func (t *Thread) handleRegisterModulesToCore(request *thread.ProvisionerRequest, response *thread.ProvisionerResponse) {

	mandatory := thread.SocketMandatory{
		Pipe:      t.channels.C0,
		NoncePool: t.noncePool,
	}
	for _, moduleInst := range t.useCases.GetModules() {
		ir := moduleInst.GetIR()
		thread.AsyncModuleAdd(mandatory, ir)
	}
}
