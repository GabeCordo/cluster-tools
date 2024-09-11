package processor

import "testing"

// TestCluster_Add
// Test that the number of processors is incremented after Add
func TestCluster_Add(t *testing.T) {

	cluster := newCluster("test")

	processor := newProcessor("localhost", 8000)
	cluster.Add(processor)

	if cluster.numOfProcessors != 1 {
		t.Error("expected the number of processors to be 1")
	}
}

// TestCluster_SelectProcessor
// Test that processors supporting a cluster get selected in a circular
// fashion so that balances are distributed equally across them.
func TestCluster_SelectProcessor(t *testing.T) {

	cluster := newCluster("test")

	processor1 := newProcessor("localhost", 8000)
	cluster.Add(processor1)
	processor2 := newProcessor("localhost", 8001)
	cluster.Add(processor2)
	processor3 := newProcessor("localhost", 8002)
	cluster.Add(processor3)
	processor4 := newProcessor("localhost", 8003)
	cluster.Add(processor4)

	processors := []*Processor{processor1, processor2, processor3, processor4}

	for i := 0; i < 8; i++ {
		expectedProcessor := processors[i%4]
		selectedProcessor := cluster.SelectProcessor()
		if selectedProcessor.Port != expectedProcessor.Port {
			t.Errorf("expected selected processor to be (%s:%d)\n",
				expectedProcessor.Host, expectedProcessor.Port)
		}
	}

}
