package provisioner

import (
	"errors"
	"fmt"
	"time"

	"github.com/GabeCordo/Flock/internal/processor/thread"
)

func (t *Thread) Setup() {

	err := t.registerModulesToCore()
	if err != nil {
		fmt.Print(err)
	}
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
				t.requestWg.Add(1)
				oRsp = t.processRequest(iReq)
				if oRsp != nil {
					t.channels.C2 <- oRsp
				}
				t.requestWg.Done()
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

func (t *Thread) registerModulesToCore() error {

	var req *thread.SocketRequest

	// logging enhancements
	for _, moduleInst := range t.provisioner.GetModules() {
		cfg := moduleInst.ToConfig()

		req = thread.NewSocketRequest()
		req.Action = thread.SocketModuleAdd
		req.Data = cfg
		req.Nonce = t.noncePool.Next()

		t.channels.C0 <- req
	}

	return nil
}

func (t *Thread) backlog() {

	for {

		if (t.NumOfActiveSupervisors() < MaxNumOfSupervisors) && (len(t.requestBacklog) > 0) {
			request := t.requestBacklog[0]
			t.channels.C1 <- request
			t.requestBacklog = t.requestBacklog[1:]
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (t *Thread) processRequest(request *thread.ProvisionerRequest) (response *thread.ProvisionerResponse) {

	response = thread.NewProvisionerResponse(request)

	switch request.Action {
	case thread.ProvisionerModuleGet:
		{
			response.Error = errors.New("implement me")
		}
	case thread.ProvisionerRunCreate:
		{
			response.Error = t.provisionRun(request)
		}
	case thread.ProvisionerRunStop:
		{
			response.Error = t.stopRun(request)
		}
	case thread.ProvisionerStatisticsGet:
		{
			response.Data = t.getStatistics()
		}
	case thread.ProvisionerRegisterModules:
		{
			response.Error = t.registerModulesToCore()
		}
	default:
		{
			response.Error = errors.New("bad request")
		}
	}

	response.Success = response.Error == nil
	return response
}

func (t *Thread) NumOfActiveSupervisors() int {
	t.backlogMutex.RLock()
	defer t.backlogMutex.RUnlock()

	return t.numOfActiveSupervisors
}

func (t *Thread) IncrementActiveSupervisors() {
	t.backlogMutex.Lock()
	defer t.backlogMutex.Unlock()

	t.numOfActiveSupervisors++
}

func (t *Thread) DecrementActiveSupervisors() {
	t.backlogMutex.Lock()
	defer t.backlogMutex.Unlock()

	t.numOfActiveSupervisors--
}

func (t *Thread) Teardown() {

	for _, s := range t.provisioner.GetSupervisors() {

		if !s.IsAlive() {
			fmt.Printf("runner is not alive %d %s\n", s.Id, s.State.ToString())
			continue
		}

		fmt.Printf("marking runner as teardown %d\n", s.Id)
		s.Teardown()
	}

	// wait for all the async messages to be processed before tearing down
	t.requestWg.Wait()

	// wait for all the running pipelines to complete before tearing down
	//
	// note: wait for the runs (after) async messages have been completed in
	// 		 case there was a pending run request in the async queue.
	t.runWg.Wait()
}
