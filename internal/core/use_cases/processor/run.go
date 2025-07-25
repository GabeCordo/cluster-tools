package processor

import (
	"errors"

	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (uc UseCases) FindCandidateProcessor(r *thread.Response) (*processor2.Processor, error) {

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
		moduleInstance, found := uc.ProcessorTable.GetModule(function.Module)
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

func (uc UseCases) UpdateProcessorAfterRunStarted(r *thread.Response) (uint64, error) {

	// TODO : support errors returned by classes
	return (r.Data).(uint64), nil
}
