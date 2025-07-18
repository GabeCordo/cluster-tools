package runner

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"strconv"
)

func (uc UseCases) GetRuns(namespace, pipeline string, identifier uint64) (rr []*run.Run) {

	f := database.Filter{
		Namespace:  namespace,
		Pipeline:   pipeline,
		Identifier: strconv.FormatUint(identifier, 10),
	}

	instances := uc.RunDatabase.Get(f)
	rr = make([]*run.Run, len(instances))
	for i, instance := range instances {
		rr[i] = instance.(*run.Run)
	}

	return rr
}

func (uc UseCases) GetRun(runId uint64) (*run.Run, error) {

	results := uc.RunDatabase.Get(database.Filter{Identifier: strconv.FormatUint(runId, 10)})
	if len(results) != 1 {
		return nil, errors.New("runner cannot be found")
	}

	stored, ok := (results[0]).(*run.Run)
	if !ok {
		return nil, errors.New("failed to cast result to *run.Run")
	}

	return stored, nil
}

func (uc UseCases) CreateRun(namespace, pipeline string, processor uint64, config *pipeline.Pipeline) (uint64, error) {

	filter := database.Filter{
		Processor: processor,
		Namespace: namespace,
		Pipeline:  pipeline,
	}
	result, err := uc.RunDatabase.Create(filter, *config)
	if err != nil {
		return 0, err
	}

	id, ok := result.(uint64)
	if !ok {
		return 0, errors.New("failed to cast result to uint64")
	}

	return id, nil
}
