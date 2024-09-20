package supervisor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/config"
	"strconv"
	"sync"
)

type SupervisorDatabase struct {
	supervisors map[uint64]*Supervisor
	counter     uint64
	mutex       sync.RWMutex
}

func NewSupervisorDatabase() *SupervisorDatabase {

	registry := new(SupervisorDatabase)
	registry.supervisors = make(map[uint64]*Supervisor)
	registry.counter = 1

	return registry
}

func (database *SupervisorDatabase) Get(filter database.Filter) []any {

	supervisors := make([]any, 0)

	id, err := strconv.ParseUint(filter.Identifier, 10, 64)
	if err != nil {
		return supervisors
	}

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	// if an id is provided we ignore the module and cluster
	if id != 0 {

		if s, found := database.supervisors[id]; found {
			supervisors = append(supervisors, s)
		}

		return supervisors
	}

	useCluster := filter.UseCluster()
	useModule := filter.UseModule()

	for _, s := range database.supervisors {

		if useCluster && (s.Cluster == filter.Cluster) && (s.Module == filter.Module) {
			supervisors = append(supervisors, s)
		} else if useModule && (s.Module == filter.Module) {
			supervisors = append(supervisors, s)
		} else if !useCluster && !useModule {
			supervisors = append(supervisors, s)
		}
	}

	return supervisors
}

func (database *SupervisorDatabase) Create(filter database.Filter, record any) (any, error) {

	database.mutex.Lock()
	defer database.mutex.Unlock()

	identifier := database.counter

	// todo : hack for now
	cfg, ok := record.(*config.Config)
	if !ok {
		return nil, errors.New("invalid record")
	}

	// todo: get this out of here, pass the pointer rather than create it here!
	s := New(identifier, filter.Processor, filter.Module, filter.Cluster, cfg)
	database.supervisors[identifier] = s

	database.counter++

	return identifier, nil
}

func (database *SupervisorDatabase) Replace(filter database.Filter, record any) error {

	panic("not implemented")
}

func (database *SupervisorDatabase) Delete(filter database.Filter) error {

	panic("not implemented")
}

func (database *SupervisorDatabase) Save(path string) error {

	panic("not implemented")
}

func (database *SupervisorDatabase) Load(path string) error {

	panic("not implemented")
}

func (database *SupervisorDatabase) Print() {

	panic("not implemented")
}
