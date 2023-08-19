package processor

import (
	"errors"
	"github.com/GabeCordo/mango-core/core/components/processor"
	"github.com/GabeCordo/mango-core/core/components/supervisor"
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango/utils"
	"math/rand"
)

func (thread *Thread) fetchSupervisor(r *common.ProcessorRequest) (*supervisor.Supervisor, error) {

	// processor -> all supervisor ids on the processor
	//	-	id
	// /module -> all supervisor ids of the module, on the processor
	//	-	id
	//	-	status?
	// processor/module/cluster -> all ids of that cluster, of the module, on the processor
	//	-	id
	//	-	status?
	//	-	num processed?
	// id -> the entire record of the supervisor
	//	-	full information

	request := common.SupervisorRequest{
		Action:      common.SupervisorFetch,
		Identifiers: r.Identifiers,
		Data:        r.Data,
	}
	thread.C13 <- request

	rsp, didTimeout := utils.SendAndWait(thread.SupervisorResponseTable, request.Nonce,
		common.GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		return nil, utils.NoResponseReceived
	}

	response := (rsp).(common.SupervisorResponse)

	return (response.Data).(*supervisor.Supervisor), nil
}

func (thread *Thread) createSupervisor(r *common.ProcessorRequest) (uint64, error) {

	// we need to pick out a processor we want to assign the work to
	moduleInstance, found := GetTableInstance().Get(r.Identifiers.Module)
	if !found {
		return 0, processor.ModuleDoesNotExist
	}

	if !moduleInstance.IsMounted() {
		return 0, processor.ModuleNotMounted
	}

	clusterInstance, found := moduleInstance.Get(r.Identifiers.Cluster)
	if !found {
		return 0, processor.ClusterDoesNotExist
	}

	if !clusterInstance.IsMounted() {
		return 0, processor.ClusterNotMounted
	}

	request := common.SupervisorRequest{
		Action:      common.SupervisorCreate,
		Identifiers: r.Identifiers, // will contain the module, cluster
		Data:        r.Data,        // will contain the metadata map[string]string
		Nonce:       rand.Uint32(),
	}

	processorInstance := clusterInstance.SelectProcessor()
	request.Identifiers.Processor = processorInstance.ToString()

	// send the request to the supervisor thread
	// the supervisor thread will:
	//	1. create a local record of the supervisor
	//	2. set the local record to the initial state
	//  3. send a provision request to the processor endpoint
	thread.C13 <- request

	rsp, didTimeout := utils.SendAndWait(thread.SupervisorResponseTable, request.Nonce,
		common.GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		return 0, utils.NoResponseReceived
	}

	response := (rsp).(common.SupervisorResponse)
	return (response.Data).(uint64), response.Error
}

func (thread *Thread) updateSupervisor(r *common.ProcessorRequest) error {

	request := common.SupervisorRequest{
		Action:      common.SupervisorUpdate,
		Identifiers: r.Identifiers,
		Data:        r.Data,
	}
	thread.C13 <- request

	rsp, didTimeout := utils.SendAndWait(thread.SupervisorResponseTable, request.Nonce,
		common.GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		// TODO : replace with real error
		return errors.New("supervisor doesn't exist")
	}

	response := (rsp).(common.SupervisorResponse)
	return response.Error
}
