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
	t.accepting = true
}

func (t *Thread) Start() {

	// INCOMING REQUESTS

	go func() {
		for request := range t.C1 {
			if !t.accepting {
				break
			}
			t.requestWg.Add(1)
			t.processRequest(&request)
			t.requestWg.Done()
		}
	}()

	// CLEARING THE PROVISIONER BACKLOG

	go t.backlog()
}

func (t *Thread) registerModulesToCore() error {

	// logging enhancements
	for _, moduleInst := range t.provisioner.GetModules() {
		cfg := moduleInst.ToConfig()

		t.C0 <- thread.SocketRequest{
			Action: thread.SocketModuleAdd,
			Data:   cfg,
			Nonce:  t.noncePool.Next(),
		}

	}

	return nil
}

func (t *Thread) backlog() {

	for {

		if (t.NumOfActiveSupervisors() < MaxNumOfSupervisors) && (len(t.requestBacklog) > 0) {
			request := t.requestBacklog[0]
			t.C1 <- request
			t.requestBacklog = t.requestBacklog[1:]
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (t *Thread) respond(response *thread.ProvisionerResponse) {

	t.C2 <- *response
}

func (t *Thread) processRequest(request *thread.ProvisionerRequest) {

	response := &thread.ProvisionerResponse{Error: nil, Nonce: request.Nonce}

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
	t.respond(response)
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
	t.accepting = false

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
