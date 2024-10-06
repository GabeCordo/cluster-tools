package provisioner

import (
	"errors"
	"fmt"
	"github.com/Sentmint/cluster-tools/internal/processor/api"
	"github.com/Sentmint/cluster-tools/internal/processor/threads"
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
		}

		thread.listenersWg.Wait()
	}()

	// CLEARING THE PROVISIONER BACKLOG

	go thread.backlog()

	thread.listenersWg.Wait()
	thread.requestWg.Wait()
}

func (thread *Thread) registerModulesToCore() error {

	// logging enhancements
	for _, moduleInst := range thread.provisioner.GetModules() {
		cfg := moduleInst.ToConfig()

		if err := api.CreateModule(*thread.Config.Core, &thread.Config.Processor, &cfg); err == nil {
			thread.logger.Printf("registered module %s to core\n", cfg.Name)
		} else {
			thread.logger.Printf("failed to register module %s to core: %v\n", cfg.Name, err)
			return err
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
		response.Error = errors.New("implement me")
	case threads.ProvisionerRunCreate:
		response.Error = thread.provisionRun(request)
	case threads.ProvisionerStatisticsGet:
		response.Data = thread.getStatistics()
	case threads.ProvisionerRegisterModules:
		response.Error = thread.registerModulesToCore()
	default:
		response.Error = errors.New("bad request")
		thread.requestWg.Done()
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

	thread.requestWg.Wait()
}
