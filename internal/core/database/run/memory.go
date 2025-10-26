package run

import (
	"errors"
	"github.com/FortifiedCode/plover"
	"strconv"
	"sync"

	"github.com/FortifiedCode/flock/internal/core/database"
)

type LocalDatabase struct {
	runs    map[uint64]*Run
	counter uint64
	mutex   sync.RWMutex
}

func NewLocalDatabase() *LocalDatabase {

	registry := new(LocalDatabase)
	registry.runs = make(map[uint64]*Run)
	registry.counter = 1

	return registry
}

// Get returns a list of *plover.PipelineIR records from the run.LocalDatabase.
func (localDatabase *LocalDatabase) Get(filter database.Filter) []any {

	supervisors := make([]any, 0)

	id, err := strconv.ParseUint(filter.Identifier, 10, 64)
	if err != nil {
		return supervisors
	}

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	// if an id is provided we ignore the module and cluster
	if id != 0 {

		if s, found := localDatabase.runs[id]; found {
			supervisors = append(supervisors, s)
		}

		return supervisors
	}

	usePipeline := filter.UsePipeline()
	useModule := filter.UseNamespace()

	for _, s := range localDatabase.runs {

		if usePipeline && (s.Pipeline.Identifier == filter.Pipeline) && (s.Namespace == filter.Namespace) {
			supervisors = append(supervisors, s)
		} else if useModule && (s.Namespace == filter.Namespace) {
			supervisors = append(supervisors, s)
		} else if !usePipeline && !useModule {
			supervisors = append(supervisors, s)
		}
	}

	return supervisors
}

// Create adds a *plover.PipelineIR record to the run.LocalDatabase.
func (localDatabase *LocalDatabase) Create(filter database.Filter, record any) (any, error) {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	identifier := localDatabase.counter

	// todo : hack for now
	cfg, ok := record.(*plover.PipelineIR)
	if !ok {
		return nil, errors.New("invalid record")
	}

	// todo: get this out of here, pass the pointer rather than create it here!
	s := New(identifier, filter.Processor, filter.Namespace, cfg)
	localDatabase.runs[identifier] = s

	localDatabase.counter++

	return identifier, nil
}

// Replace is not implemented for the run.LocalDatabase struct.
func (localDatabase *LocalDatabase) Replace(filter database.Filter, record any) (err error) {

	err = database.NotImplemented
	return err
}

// Delete is not implemented for the run.LocalDatabase struct.
func (localDatabase *LocalDatabase) Delete(filter database.Filter) (err error) {

	err = database.NotImplemented
	return err
}

// Save is not implemented for the run.LocalDatabase struct.
func (localDatabase *LocalDatabase) Save(path string) (err error) {

	err = database.NotImplemented
	return err
}

// Load is not implemented for the run.LocalDatabase struct.
func (localDatabase *LocalDatabase) Load(path string) (err error) {

	err = database.NotImplemented
	return err
}

// Print is not implemented for the run.LocalDatabase struct.
func (localDatabase *LocalDatabase) Print() {

	// nop
}
