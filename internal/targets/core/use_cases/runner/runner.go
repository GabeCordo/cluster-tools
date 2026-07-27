package runner

import (
	"errors"
	"strconv"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/run"
)

func (uc UseCases) GetRuns(namespace, pipeline string, identifier, maximum, offset uint64) (rr []*run.Run) {

	f := database.Filter{
		Namespace:       namespace,
		Pipeline:        pipeline,
		Identifier:      strconv.FormatUint(identifier, 10),
		MaximumResults:  maximum,
		OffsetOfResults: offset,
	}

	results := uc.RunDatabase.Get(f)
	rr = make([]*run.Run, len(results))
	for i, result := range results {
		rr[i] = result
	}

	return rr
}

func (uc UseCases) GetRun(runId uint64) (*run.Run, error) {

	results := uc.RunDatabase.Get(database.Filter{Identifier: strconv.FormatUint(runId, 10)})
	if len(results) != 1 {
		return nil, errors.New("runner cannot be found")
	}

	return results[0], nil
}

func (uc UseCases) CountRuns(namespace, pipeline string) (count uint32) {

	f := database.Filter{
		Namespace: namespace,
		Pipeline:  pipeline,
	}

	count = uc.RunDatabase.Count(f)
	return count
}

func (uc UseCases) CreateRun(namespace, pipeline string, processor uint64, config *ScalingFunctions.PipelineIR, startedBy run.StartedBy) (uint64, error) {

	filter := database.Filter{
		Processor: processor,
		Namespace: namespace,
		Pipeline:  pipeline,
	}

	id, err := uc.RunDatabase.Create(filter, config, startedBy)
	if err != nil {
		return 0, err
	}

	return id, nil
}
