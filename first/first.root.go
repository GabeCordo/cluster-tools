package

import (
	"github.com/GabeCordo/etl/components/cluster"
	"github.com/GabeCordo/etl/core"
)
first

import (
	"etl/utils/cli"
	"etl/components/cluster"
	"etl/core"
	"src/vector.etl"
)

// DO NOT TOUCH THIS FILE UNLESS YOU ARE CERTAIN ABOUT WHAT YOU ARE DOING

func main() {
	c := core.NewCore()

	// DEFINED CLUSTERS START

	m := Multiply{} // A structure implementing the etl.Cluster.Cluster interface
	c.Cluster("multiply", m, cluster.Config{Identifier: "multiply"})

	// DEFINED CLUSTERS END

	if commandLine, ok := cli.NewCommandLine(c); ok {
		commandLine.Run()
	}
}
