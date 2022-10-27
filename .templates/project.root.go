package <project>

import (
	"etl/utils/cli"
	"etl/components/cluster"
	"etl/core"
	"src/vector.etl"
)

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
