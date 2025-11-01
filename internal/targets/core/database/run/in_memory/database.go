package in_memory

import (
	"github.com/FortifiedCode/flock/internal/targets/core/database"
	"github.com/FortifiedCode/flock/internal/targets/core/database/run"
	"github.com/FortifiedCode/plover"
	"strconv"
	"sync"
)

type LocalDatabase struct {
	runs    map[uint64]*run.Run
	counter uint64
	mutex   sync.RWMutex
}

func NewLocalDatabase() *LocalDatabase {

	registry := new(LocalDatabase)
	registry.runs = make(map[uint64]*run.Run)
	registry.counter = 1

	return registry
}

// Get returns a list of *plover.PipelineIR records from the run.LocalDatabase.
func (localDatabase *LocalDatabase) Get(filter database.Filter) (runs []*run.Run) {

	id, err := strconv.ParseUint(filter.Identifier, 10, 64)
	if err != nil {
		return runs
	}

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	// if an id is provided we ignore the module and cluster
	if id != 0 {

		if r, found := localDatabase.runs[id]; found {
			runs = append(runs, r)
		}

		return runs
	}

	usePipeline := filter.UsePipeline()
	useModule := filter.UseNamespace()

	for _, r := range localDatabase.runs {

		if usePipeline && (r.Pipeline.Identifier == filter.Pipeline) && (r.Namespace == filter.Namespace) {
			runs = append(runs, r)
		} else if useModule && (r.Namespace == filter.Namespace) {
			runs = append(runs, r)
		} else if !usePipeline && !useModule {
			runs = append(runs, r)
		}
	}

	return runs
}

// Create adds a *plover.PipelineIR record to the run.LocalDatabase.
func (localDatabase *LocalDatabase) Create(filter database.Filter, data *plover.PipelineIR) (uint64, error) {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	identifier := localDatabase.counter

	// todo: get this out of here, pass the pointer rather than create it here!
	s := run.New(identifier, filter.Processor, filter.Namespace, data)
	localDatabase.runs[identifier] = s

	localDatabase.counter++

	return identifier, nil
}

func (localDatabase *LocalDatabase) Print() {

}
