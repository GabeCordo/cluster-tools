package supervisor

import (
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/config"
	"strconv"
	"testing"
)

var (
	ProcessorName = "test-proc"
	ModuleName    = "test-mod"
	ClusterName   = "test-mod"
)

func TestRegistry_Create(t *testing.T) {

	registry := NewSupervisorDatabase()

	filter := database.Filter{
		Processor: ProcessorName,
		Module:    ModuleName,
		Cluster:   ClusterName,
		Config:    "tmp",
	}

	cfg := &config.Config{} // todo: this is a temp hack
	id, err := registry.Create(filter, cfg)
	if err != nil {
		t.Error("failed to create a new supervisor")
		return
	}

	if id.(uint64) != 1 {
		t.Error("id should be 1")
	}
}

func TestRegistry_Get(t *testing.T) {

	registry := NewSupervisorDatabase()

	filter := database.Filter{
		Processor: ProcessorName,
		Module:    ModuleName,
		Cluster:   ClusterName,
		Config:    "tmp",
	}

	cfg := &config.Config{}
	id, err := registry.Create(filter, cfg)
	if err != nil {
		t.Error("failed to create a new supervisor")
	}

	if id.(uint64) != 1 {
		t.Error("id should be 1")
	}

	f := database.Filter{Identifier: strconv.FormatUint(id.(uint64), 10)}
	results := registry.Get(f)

	if len(results) != 1 {
		t.Error("failed to find a supervisor record that exists")
	}

	if (results[0].(*Supervisor)).Cluster != ClusterName {
		t.Error("supervisor failed to database the correct config record")
	}
}
