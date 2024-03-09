package processor

import (
	"github.com/GabeCordo/cluster-tools/core/interfaces/processor"
	"testing"
)

// TestTable_AddProcessor
// The processor shall addCluster the new config and increment its counter.
func TestTable_AddProcessor(t *testing.T) {

	table := NewTable()

	cfg := &processor.Config{Host: "127.0.0.1", Port: 1204}

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

	cfg := &processor.Config{Host: "127.0.0.1", Port: 1204}

	table.AddProcessor(cfg)

	if err := table.AddProcessor(cfg); err == nil {
		t.Error("expected the table to reject a duplicate processor")
		return
	}

	if table.NumOfProcessors != 1 {
		t.Error("expected only 1 processor to be in the table")
	}
}
