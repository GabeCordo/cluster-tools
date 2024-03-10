package supervisor

import (
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"testing"
)

var (
	ProcessorName = "test-proc"
	ModuleName    = "test-mod"
	ClusterName   = "test-mod"
)

func TestRegistry_Create(t *testing.T) {

	registry := NewRegistry()

	config := &interfaces.Config{Identifier: ProcessorName}
	if supervisorId := registry.Create(ProcessorName, ModuleName, ClusterName, config); supervisorId == 0 {
		t.Error("failed to create a new supervisor")
	}
}

func TestRegistry_Get(t *testing.T) {

	registry := NewRegistry()

	config := &interfaces.Config{Identifier: ClusterName}
	supervisorId := registry.Create(ProcessorName, ModuleName, ClusterName, config)
	if supervisorId == 0 {
		t.Error("failed to create a new supervisor")
	}

	if supervisor, found := registry.Get(supervisorId); !found {
		t.Error("failed to find a supervisor record that exists")
	} else if found && (supervisor.Config.Identifier != ClusterName) {
		t.Error("supervisor failed to store the correct config record")
	}
}
