package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/database/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/processor"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

func (t *Thread) getSupervisor(r *thread.Request) ([]*supervisor.Supervisor, error) {

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

	request := thread.Request{
		Action:      thread.GetAction,
		Type:        thread.SupervisorRecord,
		Identifiers: r.Identifiers,
		Nonce:       r.Nonce,
	}
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.SupervisorResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(thread.Response)

	return (response.Data).([]*supervisor.Supervisor), nil
}

func (t *Thread) createSupervisor(r *thread.Request) (uint64, error) {

	// we need to pick out a processor we want to assign the work to
	moduleInstance, found := t.processorTable.GetModule(r.Identifiers.Module)
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

	if (r.Source == thread.HttpClient) && clusterInstance.IsStream() {
		return 0, processor.CanNotProvisionStreamCluster
	}

	request := thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.SupervisorRecord,
		Identifiers: r.Identifiers, // will contain the module, cluster
		Caller:      thread.User,
		Data:        r.Data, // will contain the metadata map[string]string
		Nonce:       rand.Uint32(),
	}

	processorInstance := clusterInstance.SelectProcessor()
	request.Identifiers.Processor = processorInstance.ToString()

	// send the request to the supervisor t
	// the supervisor t will:
	//	1. create a log record of the supervisor
	//	2. set the log record to the initial state
	//  3. send a provision request to the processor endpoint
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.SupervisorResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		return 0, multithreaded.NoResponseReceived
	}

	response := (rsp).(thread.Response)
	return (response.Data).(uint64), response.Error
}

func (t *Thread) updateSupervisor(r *thread.Request) error {

	request := thread.Request{
		Action:      thread.UpdateAction,
		Type:        thread.SupervisorRecord,
		Identifiers: r.Identifiers,
		Data:        r.Data,
		Nonce:       rand.Uint32(),
	}
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.SupervisorResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		// TODO : replace with real error
		return errors.New("supervisor doesn't exist")
	}

	response := (rsp).(thread.Response)
	return response.Error
}

func (t *Thread) logSupervisor(r *thread.Request) error {

	request := thread.Request{
		Action:      thread.LogAction,
		Type:        thread.SupervisorRecord,
		Identifiers: r.Identifiers,
		Data:        r.Data,
		Nonce:       rand.Uint32(),
	}
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.SupervisorResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(thread.Response)
	return response.Error
}
