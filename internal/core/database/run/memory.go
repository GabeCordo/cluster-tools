package run

import (
	"errors"
	"strconv"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
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

func (database *LocalDatabase) Get(filter database.Filter) []any {

	supervisors := make([]any, 0)

	id, err := strconv.ParseUint(filter.Identifier, 10, 64)
	if err != nil {
		return supervisors
	}

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	// if an id is provided we ignore the module and cluster
	if id != 0 {

		if s, found := database.runs[id]; found {
			supervisors = append(supervisors, s)
		}

		return supervisors
	}

	usePipeline := filter.UsePipeline()
	useModule := filter.UseNamespace()

	for _, s := range database.runs {

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

func (database *LocalDatabase) Create(filter database.Filter, record any) (any, error) {

	database.mutex.Lock()
	defer database.mutex.Unlock()

	identifier := database.counter

	// todo : hack for now
	cfg, ok := record.(*pipeline.Pipeline)
	if !ok {
		return nil, errors.New("invalid record")
	}

	// todo: get this out of here, pass the pointer rather than create it here!
	s := New(identifier, filter.Processor, filter.Namespace, cfg)
	database.runs[identifier] = s

	database.counter++

	return identifier, nil
}

func (database *LocalDatabase) Replace(filter database.Filter, record any) error {

	panic("not implemented")
}

func (database *LocalDatabase) Delete(filter database.Filter) error {

	panic("not implemented")
}

func (database *LocalDatabase) Save(path string) error {

	panic("not implemented")
}

func (database *LocalDatabase) Load(path string) error {

	panic("not implemented")
}

func (database *LocalDatabase) Print() {

	panic("not implemented")
}
