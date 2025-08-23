package provisioner

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/processor/use_cases/provisioner"
	"time"

	"github.com/FortifiedCode/flock/internal/processor/thread"
)

func (t *Thread) Setup() {

}

func (t *Thread) Start() {

	go t.backlog()

	var iReq *thread.ProvisionerRequest
	var oRsp *thread.ProvisionerResponse
	stop := false

	for {
		select {
		case iReq = <-t.channels.C1:
			{
				oRsp = t.processRequest(iReq)
				if oRsp != nil {
					t.channels.C2 <- oRsp
				}
			}
		case <-t.channels.close:
			{
				stop = true
			}
		}

		oRsp = nil

		if stop {
			break
		}
	}
}

func (t *Thread) backlog() {

	for {

		t.useCases.CheckBacklog(func(request *provisioner.ProvisionRequest) {
			pReq := thread.NewProvisionerRequest()
			pReq.Action = thread.ProvisionerRunCreate
			pReq.Namespace = request.Namespace
			pReq.Supervisor = request.Supervisor
			pReq.Metadata = request.Metadata
			pReq.Pipeline = request.Pipeline
			t.channels.C1 <- pReq
		})

		time.Sleep(10 * time.Millisecond)
	}
}

func (t *Thread) processRequest(request *thread.ProvisionerRequest) (response *thread.ProvisionerResponse) {

	response = thread.NewProvisionerResponse(request)

	switch request.Action {
	case thread.ProvisionerModuleGet:
		{
			t.handleGetModules(request, response)
		}
	case thread.ProvisionerRunCreate:
		{
			t.handleCreateRun(request, response)
		}
	case thread.ProvisionerRunStop:
		{
			t.handleStopRun(request, response)
		}
	case thread.ProvisionerStatisticsGet:
		{
			t.handleGetStatistics(request, response)
		}
	case thread.ProvisionerRegisterModules:
		{
			t.handleRegisterModulesToCore(request, response)
		}
	default:
		{
			response.Error = errors.New("bad request")
		}
	}

	response.Success = response.Error == nil
	return response
}

func (t *Thread) Teardown() {

	t.useCases.StopAllRuns()
}
