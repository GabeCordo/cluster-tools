package in_memory

import (
	"strconv"
	"testing"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/run"
)

var (
	ProcessorName = 0
	ModuleName    = "test-mod"
	ClusterName   = "test-mod"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := NewLocalDatabase()

	var i run.Database
	i = d

	i.Print()
}

func TestRegistry_Create(t *testing.T) {

	registry := NewLocalDatabase()

	filter := database.Filter{
		Processor: uint64(ProcessorName),
		Namespace: ModuleName,
		Pipeline:  ClusterName,
		Config:    "tmp",
	}

	cfg := &ScalingFunctions.PipelineIR{} // todo: this is a temp hack
	id, err := registry.Create(filter, cfg, run.Operator)
	if err != nil {
		t.Error("failed to create a new runner")
		return
	}

	if id != 1 {
		t.Error("id should be 1")
	}
}

func TestRegistry_Get(t *testing.T) {

	registry := NewLocalDatabase()

	filter := database.Filter{
		Processor: uint64(ProcessorName),
		Namespace: ModuleName,
		Pipeline:  ClusterName,
		Config:    "tmp",
	}

	cfg := &ScalingFunctions.PipelineIR{Identifier: ClusterName}
	id, err := registry.Create(filter, cfg, run.Operator)
	if err != nil {
		t.Error("failed to create a new runner")
	}

	if id != 1 {
		t.Error("id should be 1")
	}

	f := database.Filter{Identifier: strconv.FormatUint(id, 10)}
	results := registry.Get(f)

	if len(results) != 1 {
		t.Error("failed to find a runner record that exists")
	}

	if results[0].Pipeline.Identifier != ClusterName {
		t.Error("runner failed to database the correct pipeline record")
	}
}
