package processor

import (
	"github.com/GabeCordo/plover"
	"testing"
)

// TestCluster_Add
// Test that the number of Processors is incremented after Add
func TestCluster_Add(t *testing.T) {

	function := newFunction(&plover.FunctionIR{Identifier: "test"})

	processor := newProcessor(0, "localhost:8000")
	function.Add(processor)

	if function.numOfProcessors != 1 {
		t.Error("expected the number of Processors to be 1")
	}
}

// TestCluster_SelectProcessor
// Test that Processors supporting a cluster get selected in a circular
// fashion so that balances are distributed equally across them.
func TestCluster_SelectProcessor(t *testing.T) {

	function := newFunction(&plover.FunctionIR{Identifier: "test"})

	processor1 := newProcessor(0, "localhost:8000")
	function.Add(processor1)
	processor2 := newProcessor(1, "localhost:8001")
	function.Add(processor2)
	processor3 := newProcessor(2, "localhost:8002")
	function.Add(processor3)
	processor4 := newProcessor(3, "localhost:8003")
	function.Add(processor4)

	processors := []*Processor{processor1, processor2, processor3, processor4}

	for i := 0; i < 8; i++ {
		expectedProcessor := processors[i%4]
		selectedProcessor := function.SelectProcessor()
		if selectedProcessor.Id != expectedProcessor.Id {
			t.Errorf("expected selected processor to be (%s)\n",
				expectedProcessor.RemoteAddr)
		}
	}

}
