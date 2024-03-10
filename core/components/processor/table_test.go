package processor

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"testing"
)

// TestTable_AddProcessor
// The processor shall addCluster the new config and increment its counter.
func TestTable_AddProcessor(t *testing.T) {

	table := NewTable()

	cfg := &interfaces.ProcessorConfig{Host: "127.0.0.1", Port: 1204}

	if err := table.AddProcessor(cfg); err != nil {
		t.Error(err)
		return
	}

	if table.NumOfProcessors != 1 {
		t.Error("num of processors not incremented correctly")
	}
}

// TestTable_AddProcessor2
// The table should reject a processor if its address and port already
// exist as a 1:1 match in the table with no change to its counter.
func TestTable_AddProcessor2(t *testing.T) {

	table := NewTable()

	cfg := &interfaces.ProcessorConfig{Host: "127.0.0.1", Port: 1204}

	table.AddProcessor(cfg)

	if err := table.AddProcessor(cfg); err == nil {
		t.Error("expected the table to reject a duplicate processor")
		return
	}

	if table.NumOfProcessors != 1 {
		t.Error("expected only 1 processor to be in the table")
	}
}

// TestTable_AddModule
// Add a module for a processor that doesn't exist
func TestTable_AddModule(t *testing.T) {

	table := NewTable()

	moduleConfig := &interfaces.ModuleConfig{}
	if err := table.AddModule("foo", moduleConfig); !errors.Is(err, DoesNotExist) {
		t.Error("table should throw DoesNotExist for unknown processor")
	}
}

// TestTable_AddModule2
// Add a module for a processor that does exist, then add a second processor with the same module
// but verify the module object is not changed
func TestTable_AddModule2(t *testing.T) {

	table := NewTable()

	processorConfig := &interfaces.ProcessorConfig{Host: "127.0.0.1", Port: 1204}
	table.AddProcessor(processorConfig)

	moduleConfig := &interfaces.ModuleConfig{Name: "foo", Exports: make([]interfaces.ModuleCluster, 1)}
	moduleConfig.Exports[0] = interfaces.ModuleCluster{"bar", false, interfaces.ModuleClusterConfig{}}

	if err := table.AddModule("127.0.0.1:1204", moduleConfig); err != nil {
		t.Error(err)
		return
	}

	moduleInstance, found := table.GetModule("foo")
	if !found {
		t.Error("expected to find module instance")
		return
	}

	clusterInstance, found := moduleInstance.GetCluster("bar")
	if !found {
		t.Error("expected to find cluster instance under module")
		return
	}

	if clusterInstance.numOfProcessors != 1 {
		t.Error("expected cluster record to have 1 processor")
		return
	}

	if clusterInstance.SelectProcessor().ToString() != "127.0.0.1:1204" {
		t.Error("no processor record found under module.config")
		return
	}

	processorConfig2 := &interfaces.ProcessorConfig{Host: "127.0.0.1", Port: 1205}
	table.AddProcessor(processorConfig2)

	if err := table.AddModule("127.0.0.1:1205", moduleConfig); err != nil {
		t.Error(err)
		return
	}

	if fetchedModule, _ := table.GetModule("foo"); fetchedModule != moduleInstance {
		t.Error("module object should not be changed")
		return
	}

	if clusterInstance.numOfProcessors != 2 {
		t.Error("expected the cluster to now have two supporting processors")
	}
}
