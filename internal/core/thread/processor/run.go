package processor

import (
	"errors"
	"github.com/GabeCordo/toolchain/multithreaded"
	"github.com/Sentmint/PipelineOps/internal/core/database/run"
	"github.com/Sentmint/PipelineOps/internal/core/processor"
	"github.com/Sentmint/PipelineOps/internal/core/thread"
	"math/rand"
)

func (t *Thread) getRun(r *thread.Request) ([]*run.Run, error) {

	// processor -> all runner ids on the processor
	//	-	id
	// /module -> all runner ids of the module, on the processor
	//	-	id
	//	-	status?
	// processor/module/cluster -> all ids of that cluster, of the module, on the processor
	//	-	id
	//	-	status?
	//	-	num processed?
	// id -> the entire record of the runner
	//	-	full information

	request := thread.Request{
		Action:      thread.GetAction,
		Type:        thread.RunRecord,
		Identifiers: r.Identifiers,
		Nonce:       r.Nonce,
	}
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.RunnerResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(thread.Response)

	return (response.Data).([]*run.Run), nil
}

func (t *Thread) createRun(r *thread.Request) (uint64, error) {

	mandatory := thread.Mandatory{
		Pipe:          t.C11,
		ResponseTable: t.DatabaseResponseTable,
		Timeout:       t.config.Timeout,
	}
	pipelineDescription, found := thread.GetPipelineFromDatabase(mandatory, r.Identifiers.Namespace, r.Identifiers.Pipeline)
	if !found {
		return 0, errors.New("pipeline not found")
	}

	processors := make(map[string]*processor.Processor)

	// validate that each functions module exists
	for _, function := range pipelineDescription.Functions {

		// we need to pick out a processor we want to assign the work to
		moduleInstance, found := t.processorTable.GetModule(function.Module)
		if !found {
			return 0, processor.ModuleDoesNotExist
		}

		if !moduleInstance.IsMounted() {
			return 0, processor.ModuleNotMounted
		}

		functionInstance, found := moduleInstance.GetFunction(function.Identifier)
		if !found {
			return 0, processor.FunctionDoesNotExist
		}

		if !functionInstance.IsMounted() {
			return 0, processor.FunctionNotMounted
		}

		for _, p := range functionInstance.Processors {
			processors[p.ToString()] = p
		}
	}

	// TODO: remove?
	//if (r.Source == thread.HttpClient) && functionInstance.IsStream() {
	//	return 0, processor.CanNotProvisionStreamCluster
	//}

	request := thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.RunRecord,
		Identifiers: r.Identifiers, // will contain the module, cluster
		Caller:      thread.User,
		Data:        r.Data, // will contain the metadata map[string]string
		Nonce:       rand.Uint32(),
	}

	// if there are no viable processors, stop
	if len(processors) == 0 {
		return 0, errors.New("no processors are available to support the pipeline")
	}

	// select one of the processors
	var selectedProcessor *processor.Processor = nil
	for _, p := range processors {

		if selectedProcessor == nil {
			selectedProcessor = p
		} else if p.NumOfRuns < selectedProcessor.NumOfRuns {
			selectedProcessor = p
		}
	}
	// unlikely
	if selectedProcessor == nil {
		return 0, errors.New("no processors are available to support the pipeline")
	}
	selectedProcessor.NumOfRuns++
	request.Identifiers.Processor = selectedProcessor.Id

	// send the request to the scheduler t
	// the scheduler t will:
	//	1. create a log record of the runner
	//	2. set the log record to the initial state
	//  3. send a provision request to the processor endpoint
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.RunnerResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		return 0, multithreaded.NoResponseReceived
	}

	response := (rsp).(thread.Response)
	return (response.Data).(uint64), response.Error
}

func (t *Thread) updateRun(r *thread.Request) error {

	request := thread.Request{
		Action:      thread.UpdateAction,
		Type:        thread.RunRecord,
		Identifiers: r.Identifiers,
		Data:        r.Data,
		Nonce:       rand.Uint32(),
	}
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.RunnerResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		// TODO : replace with real error
		return errors.New("runner doesn't exist")
	}

	response := (rsp).(thread.Response)
	return response.Error
}

func (t *Thread) logRun(r *thread.Request) error {

	request := thread.Request{
		Action:      thread.LogAction,
		Type:        thread.RunRecord,
		Identifiers: r.Identifiers,
		Data:        r.Data,
		Nonce:       rand.Uint32(),
	}
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.RunnerResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(thread.Response)
	return response.Error
}

func (t *Thread) stopRun(r *thread.Request) error {

	request := thread.Request{
		Action:      thread.DeleteAction,
		Type:        thread.RunRecord,
		Identifiers: r.Identifiers,
		Nonce:       rand.Uint32(),
	}
	t.C13 <- request

	rsp, didTimeout := multithreaded.SendAndWait(t.RunnerResponseTable, request.Nonce,
		t.config.Timeout)

	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(thread.Response)
	return response.Error
}
