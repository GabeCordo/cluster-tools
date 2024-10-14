package provisioner

import (
	"errors"
	"fmt"
	"github.com/Sentmint/cluster-tools/internal/processor/threads"
	"math/rand"
	"time"
)

func (thread *Thread) Setup() {

	thread.registerModulesToCore()
	thread.accepting = true
}

func (thread *Thread) Start() {

	// INCOMING REQUESTS

	go func() {
		for request := range thread.C1 {
			if !thread.accepting {
				break
			}
			thread.requestWg.Add(1)
			thread.processRequest(&request)
			thread.requestWg.Done()
		}
	}()

	// CLEARING THE PROVISIONER BACKLOG

	go thread.backlog()
}

func (thread *Thread) registerModulesToCore() error {

	// logging enhancements
	for _, moduleInst := range thread.provisioner.GetModules() {
		cfg := moduleInst.ToConfig()

		thread.C0 <- threads.SocketRequest{
			Action: threads.SocketModuleAdd,
			Data:   cfg,
			Nonce:  rand.Uint32(),
		}

	}

	return nil
}

func (thread *Thread) backlog() {

	for {

		if (thread.NumOfActiveSupervisors() < MaxNumOfSupervisors) && (len(thread.requestBacklog) > 0) {
			request := thread.requestBacklog[0]
			thread.C1 <- request
			thread.requestBacklog = thread.requestBacklog[1:]
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (thread *Thread) respond(response *threads.ProvisionerResponse) {

	thread.C2 <- *response
}

func (thread *Thread) processRequest(request *threads.ProvisionerRequest) {

	response := &threads.ProvisionerResponse{Error: nil, Nonce: request.Nonce}

	switch request.Action {
	case threads.ProvisionerModuleGet:
		{
			response.Error = errors.New("implement me")
		}
	case threads.ProvisionerRunCreate:
		{
			response.Error = thread.provisionRun(request)
		}
	case threads.ProvisionerRunStop:
		{
			response.Error = thread.stopRun(request)
		}
	case threads.ProvisionerStatisticsGet:
		{
			response.Data = thread.getStatistics()
		}
	case threads.ProvisionerRegisterModules:
		{
			response.Error = thread.registerModulesToCore()
		}
	default:
		{
			response.Error = errors.New("bad request")
		}
	}

	response.Success = response.Error == nil
	thread.respond(response)
}

func (thread *Thread) NumOfActiveSupervisors() int {
	thread.backlogMutex.RLock()
	defer thread.backlogMutex.RUnlock()

	return thread.numOfActiveSupervisors
}

func (thread *Thread) IncrementActiveSupervisors() {
	thread.backlogMutex.Lock()
	defer thread.backlogMutex.Unlock()

	thread.numOfActiveSupervisors++
}

func (thread *Thread) DecrementActiveSupervisors() {
	thread.backlogMutex.Lock()
	defer thread.backlogMutex.Unlock()

	thread.numOfActiveSupervisors--
}

func (thread *Thread) Teardown() {
	thread.accepting = false

	for _, s := range thread.provisioner.GetSupervisors() {

		if !s.IsAlive() {
			fmt.Printf("runner is not alive %d %s\n", s.Id, s.State.ToString())
			continue
		}

		fmt.Printf("marking runner as teardown %d\n", s.Id)
		s.Teardown()
	}

	// wait for all the async messages to be processed before tearing down
	thread.requestWg.Wait()

	// wait for all the running pipelines to complete before tearing down
	//
	// note: wait for the runs (after) async messages have been completed in
	// 		 case there was a pending run request in the async queue.
	thread.runWg.Wait()
}
