package run

import (
	"github.com/Sentmint/PipelineOps/internal/core/database"
	"github.com/Sentmint/PipelineOps/internal/core/database/pipeline"
	"strconv"
	"testing"
)

var (
	ProcessorName = 0
	ModuleName    = "test-mod"
	ClusterName   = "test-mod"
)

func TestRegistry_Create(t *testing.T) {

	registry := NewLocalDatabase()

	filter := database.Filter{
		Processor: uint64(ProcessorName),
		Namespace: ModuleName,
		Pipeline:  ClusterName,
		Config:    "tmp",
	}

	cfg := &pipeline.Pipeline{} // todo: this is a temp hack
	id, err := registry.Create(filter, cfg)
	if err != nil {
		t.Error("failed to create a new runner")
		return
	}

	if id.(uint64) != 1 {
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

	cfg := &pipeline.Pipeline{Identifier: ClusterName}
	id, err := registry.Create(filter, cfg)
	if err != nil {
		t.Error("failed to create a new runner")
	}

	if id.(uint64) != 1 {
		t.Error("id should be 1")
	}

	f := database.Filter{Identifier: strconv.FormatUint(id.(uint64), 10)}
	results := registry.Get(f)

	if len(results) != 1 {
		t.Error("failed to find a runner record that exists")
	}

	if (results[0].(*Run)).Pipeline.Identifier != ClusterName {
		t.Error("runner failed to database the correct pipeline record")
	}
}
