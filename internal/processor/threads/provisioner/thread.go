package provisioner

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/clarence/internal/api"
	"github.com/GabeCordo/clarence/internal/threads/common"
	"time"
)

func (thread *Thread) Setup() {

	thread.accepting = true

	// initialize a modules instance with a common module
	GetProvisionerInstance()
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

	// STANDALONE MODE

	if thread.Config.Standalone {
		thread.provisionStreamFunction()
	} else {
		thread.registerModulesToCore()
	}

	// CLEARING THE PROVISIONER BACKLOG

	go thread.backlog()

	thread.listenersWg.Wait()
	thread.requestWg.Wait()
}

func (thread *Thread) provisionStreamFunction() {
	for _, moduleInst := range GetProvisionerInstance().GetModules() {

		for _, clusterInst := range moduleInst.GetClusters() {

			if !clusterInst.IsStream() {
				continue
			}

			request := &common.ProvisionerRequest{
				Action:   common.ProvisionerSupervisorCreate,
				Source:   common.Core,
				Module:   moduleInst.Identifier,
				Cluster:  clusterInst.Identifier,
				Config:   &clusterInst.DefaultConfig,
				Metadata: make(map[string]string),
			}
			thread.requestWg.Add(1)
			thread.provisionSupervisor(request)
		}
	}
}

func (thread *Thread) registerModulesToCore() {
	// logging enhancements
	for _, moduleInst := range GetProvisionerInstance().GetModules() {
		if err := api.CreateModule(thread.Config.Core, &thread.Config.Processor, moduleInst.ToConfig()); err == nil {
			thread.logger.Printf("registered module %s to core\n", moduleInst.Identifier)
		} else {
			thread.logger.Printf("failed to register module %s to core: %v\n", moduleInst.Identifier, err)
		}
	}
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

func (thread *Thread) respond(response *common.ProvisionerResponse) {

	thread.C2 <- *response
}

func (thread *Thread) processRequest(request *common.ProvisionerRequest) {

	response := &common.ProvisionerResponse{Error: nil, Nonce: request.Nonce}

	switch request.Action {
	case common.ProvisionerModuleGet:
		response.Data = thread.getModules()
	case common.ProvisionerSupervisorCreate:
		response.Error = thread.provisionSupervisor(request)
	case common.ProvisionerStatisticsGet:
		response.Data = thread.getStatistics()
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

	modules := GetProvisionerInstance().GetModules()

	for _, module := range modules {

		clusters := module.GetClusters()

		for _, cluster := range clusters {

			if !cluster.IsStream() {
				continue
			}

			supervisors := cluster.FindSupervisors()

			for _, supervisor := range supervisors {

				if !supervisor.IsAlive() {

					fmt.Printf("supervisor is not alive %d %s\n", supervisor.Id, supervisor.State.ToString())
					continue
				}

				fmt.Printf("marking supervisor as teardown %d\n", supervisor.Id)
				supervisor.Teardown()
			}
		}
	}

	thread.requestWg.Wait()
}
