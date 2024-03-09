package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/components/processor"
	"github.com/GabeCordo/cluster-tools/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

func (thread *Thread) getSupervisor(r *common.ProcessorRequest) (*supervisor.Supervisor, error) {

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
		Action:      common.SupervisorGet,
		Identifiers: r.Identifiers,
		Nonce:       r.Nonce,
	}
	thread.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(thread.SupervisorResponseTable, request.Nonce,
		thread.config.Timeout)

	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(common.SupervisorResponse)

	return (response.Data).(*supervisor.Supervisor), nil
}

func (thread *Thread) createSupervisor(r *common.ProcessorRequest) (uint64, error) {

	// we need to pick out a processor we want to assign the work to
	moduleInstance, found := GetTableInstance().GetModule(r.Identifiers.Module)
	if !found {
		return 0, processor.ModuleDoesNotExist
	}

	if !moduleInstance.IsMounted() {
		return 0, processor.ModuleNotMounted
	}

	clusterInstance, found := moduleInstance.GetCluster(r.Identifiers.Cluster)
	if !found {
		return 0, processor.ClusterDoesNotExist
	}

	if !clusterInstance.IsMounted() {
		return 0, processor.ClusterNotMounted
	}

	if (r.Source == common.HttpClient) && clusterInstance.IsStream() {
		return 0, processor.CanNotProvisionStreamCluster
	}

	request := common.SupervisorRequest{
		Action:      common.SupervisorCreate,
		Identifiers: r.Identifiers, // will contain the module, cluster
		Caller:      common.User,
		Data:        r.Data, // will contain the metadata map[string]string
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

	rsp, didTimeout := multithreaded.SendAndWait(thread.SupervisorResponseTable, request.Nonce,
		thread.config.Timeout)

	if didTimeout {
		return 0, multithreaded.NoResponseReceived
	}

	response := (rsp).(common.SupervisorResponse)
	return (response.Data).(uint64), response.Error
}

func (thread *Thread) updateSupervisor(r *common.ProcessorRequest) error {

	request := common.SupervisorRequest{
		Action:      common.SupervisorUpdate,
		Identifiers: r.Identifiers,
		Data:        r.Data,
		Nonce:       rand.Uint32(),
	}
	thread.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(thread.SupervisorResponseTable, request.Nonce,
		thread.config.Timeout)

	if didTimeout {
		// TODO : replace with real error
		return errors.New("supervisor doesn't exist")
	}

	response := (rsp).(common.SupervisorResponse)
	return response.Error
}

func (thread *Thread) logSupervisor(r *common.ProcessorRequest) error {

	request := common.SupervisorRequest{
		Action:      common.SupervisorLog,
		Identifiers: r.Identifiers,
		Data:        r.Data,
		Nonce:       rand.Uint32(),
	}
	thread.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(thread.SupervisorResponseTable, request.Nonce,
		thread.config.Timeout)

	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(common.SupervisorResponse)
	return response.Error
}
