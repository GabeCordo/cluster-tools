package in_memory

import (
	"github.com/GabeCordo/ScalingFunctions"
	"strconv"
	"sync"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/run"
)

type LocalDatabase struct {
	runs    []*run.Run
	index   map[uint64]int
	offset  int
	counter uint64
	mutex   sync.RWMutex
}

func NewLocalDatabase() *LocalDatabase {

	registry := new(LocalDatabase)
	registry.index = make(map[uint64]int)
	registry.runs = make([]*run.Run, 0)
	registry.offset = 0
	registry.counter = 1

	return registry
}

// Get returns a list of *ScalingFunctions.PipelineIR records from the run.LocalDatabase.
func (localDatabase *LocalDatabase) Get(filter database.Filter) (runs []*run.Run) {

	id, err := strconv.ParseUint(filter.Identifier, 10, 64)
	if err != nil {
		return runs
	}

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	// if an id is provided we ignore the module and cluster
	if id != 0 {
		recordIndex := localDatabase.index[id]
		runs = append(runs, localDatabase.runs[recordIndex])
		return runs
	}

	usePipeline := filter.UsePipeline()
	useModule := filter.UseNamespace()

	var r *run.Run
	var rOffset uint64 = 0

	for i := len(localDatabase.runs) - 1; i >= 0; i-- {

		r = localDatabase.runs[i]

		// skip over a series of elements that the callee may want to ignore
		if filter.UseOffset() && (rOffset < filter.OffsetOfResults) {
			rOffset++
			continue
		}

		// limit the number of results if the callee specifies
		if filter.UseMaximum() && ((rOffset - filter.OffsetOfResults) > (filter.MaximumResults - 1)) {
			break
		}

		if usePipeline && (r.Pipeline.Identifier == filter.Pipeline) && (r.Namespace == filter.Namespace) {
			runs = append(runs, r)
		} else if useModule && (r.Namespace == filter.Namespace) {
			runs = append(runs, r)
		} else if !usePipeline && !useModule {
			runs = append(runs, r)
		}

		rOffset++
	}

	return runs
}

// Count returns the number of *ScalingFunctions.PipelineIR records from the run.LocalDatabase.
func (localDatabase *LocalDatabase) Count(filter database.Filter) (count uint32) {

	count = 0

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	usePipeline := filter.UsePipeline()
	useModule := filter.UseNamespace()

	var r *run.Run
	var rOffset uint64 = 0

	for i := len(localDatabase.runs) - 1; i >= 0; i-- {

		r = localDatabase.runs[i]

		// skip over a series of elements that the callee may want to ignore
		if filter.UseOffset() && (rOffset < filter.OffsetOfResults) {
			rOffset++
			continue
		}

		// limit the number of results if the callee specifies
		if filter.UseMaximum() && ((rOffset - filter.OffsetOfResults) > (filter.MaximumResults - 1)) {
			break
		}

		if usePipeline && (r.Pipeline.Identifier == filter.Pipeline) && (r.Namespace == filter.Namespace) {
			count++
		} else if useModule && (r.Namespace == filter.Namespace) {
			count++
		} else if !usePipeline && !useModule {
			count++
		}

		rOffset++
	}

	return count
}

// Create adds a *ScalingFunctions.PipelineIR record to the run.LocalDatabase.
func (localDatabase *LocalDatabase) Create(filter database.Filter, data *ScalingFunctions.PipelineIR, startedBy run.StartedBy) (uint64, error) {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	identifier := localDatabase.counter

	// todo: get this out of here, pass the pointer rather than create it here!
	s := run.New(identifier, filter.Processor, filter.Namespace, startedBy, data)
	localDatabase.runs = append(localDatabase.runs, s)

	localDatabase.index[identifier] = localDatabase.offset

	localDatabase.offset++
	localDatabase.counter++

	return identifier, nil
}

func (localDatabase *LocalDatabase) Print() {

}
