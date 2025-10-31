package runner

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/core/database/run"
	"github.com/FortifiedCode/plover"
	"strconv"
)

func (uc UseCases) GetRuns(namespace, pipeline string, identifier uint64) (rr []*run.Run) {

	f := database.Filter{
		Namespace:  namespace,
		Pipeline:   pipeline,
		Identifier: strconv.FormatUint(identifier, 10),
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

func (uc UseCases) CreateRun(namespace, pipeline string, processor uint64, config *plover.PipelineIR) (uint64, error) {

	filter := database.Filter{
		Processor: processor,
		Namespace: namespace,
		Pipeline:  pipeline,
	}

	id, err := uc.RunDatabase.Create(filter, config)
	if err != nil {
		return 0, err
	}

	return id, nil
}
