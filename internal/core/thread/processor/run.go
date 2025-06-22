package processor

import (
	"errors"

	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) asyncGetRunFromRunner(r *thread.Request) {

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

	request := new(thread.Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = thread.GetAction
	request.Type = thread.RunRecord
	request.Identifiers = r.Identifiers
	request.Source = thread.Processor
	request.Nonce = r.Nonce

	t.channels.c13 <- request
}

func (t *Thread) asyncGetPipelineFromDatabase(r *thread.Request) {

	databaseRequest := new(thread.Request)
	if databaseRequest == nil {
		panic("failed to allocate thread.Request")
	}

	databaseRequest.Action = thread.GetAction
	databaseRequest.Type = thread.PipelineRecord
	databaseRequest.Identifiers = r.Identifiers
	databaseRequest.Source = thread.Processor
	databaseRequest.Nonce = r.Nonce

	t.channels.c11 <- databaseRequest
}

func (t *Thread) syncFindCandidateProcessor(r *thread.Response) (*processor2.Processor, error) {

	pp, ok := r.Data.([]pipeline.Pipeline)
	if !ok {
		return nil, errors.New("received invalid response from database")
	}

	if len(pp) < 1 {
		return nil, errors.New("unknown pipeline")
	}
	p := pp[0]

	processors := make(map[string]*processor2.Processor)

	// validate that each functions module exists
	for _, function := range p.Functions {

		// we need to pick out a processor we want to assign the work to
		moduleInstance, found := t.processorTable.GetModule(function.Module)
		if !found {
			return nil, processor2.ModuleDoesNotExist
		}

		if !moduleInstance.IsMounted() {
			return nil, processor2.ModuleNotMounted
		}

		functionInstance, found := moduleInstance.GetFunction(function.Identifier)
		if !found {
			return nil, processor2.FunctionDoesNotExist
		}

		if !functionInstance.IsMounted() {
			return nil, processor2.FunctionNotMounted
		}

		for _, p := range functionInstance.Processors {
			processors[p.ToString()] = p
		}
	}

	// if there are no viable processors, stop
	if len(processors) == 0 {
		return nil, errors.New("no processors are available to support the pipeline")
	}

	// select one of the processors
	var selectedProcessor *processor2.Processor = nil
	for _, p := range processors {

		if selectedProcessor == nil {
			selectedProcessor = p
		} else if p.NumOfRuns < selectedProcessor.NumOfRuns {
			selectedProcessor = p
		}
	}

	// unlikely
	if selectedProcessor == nil {
		return nil, errors.New("no processors are available to support the pipeline")
	}
	selectedProcessor.NumOfRuns++
	return selectedProcessor, nil
}

func (t *Thread) asyncSendCreateRunToRunner(processor *processor2.Processor, r *thread.Request) {

	request := new(thread.Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = thread.CreateAction
	request.Type = thread.RunRecord
	request.Identifiers = r.Identifiers // will contain the module, cluster
	request.Identifiers.Processor = processor.Id
	request.Caller = thread.User
	request.Data = r.Data // will contain the metadata map[string]string
	request.Source = thread.Processor
	request.Nonce = r.Nonce

	// send the request to the scheduler t
	// the scheduler t will:
	//	1. create a log record of the runner
	//	2. set the log record to the initial state
	//  3. send a provision request to the processor endpoint
	t.channels.c13 <- request
}

func (t *Thread) syncUpdateProcessorAfterRunStarted(r *thread.Response) (uint64, error) {

	// TODO : support errors returned by classes
	return (r.Data).(uint64), nil
}

func (t *Thread) asyncSendUpdateToRunner(r *thread.Request) {

	request := new(thread.Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = thread.UpdateAction
	request.Type = thread.RunRecord
	request.Identifiers = r.Identifiers
	request.Data = r.Data
	request.Source = thread.Processor
	request.Nonce = r.Nonce

	t.channels.c13 <- request
}

func (t *Thread) asyncSendLogToRunner(r *thread.Request) {

	request := new(thread.Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = thread.LogAction
	request.Type = thread.RunRecord
	request.Identifiers = r.Identifiers
	request.Data = r.Data
	request.Source = thread.Processor
	request.Nonce = r.Nonce

	t.channels.c13 <- request
}

func (t *Thread) asyncTellRunnerToStopRun(r *thread.Request) {

	request := new(thread.Request)

	request.Action = thread.DeleteAction
	request.Type = thread.RunRecord
	request.Identifiers = r.Identifiers
	request.Source = thread.Processor
	request.Nonce = r.Nonce

	t.channels.c13 <- request
}

func (t *Thread) syncCheckIfRunStopped(r *thread.Response) error {
	return nil
}
