package main

import (
	"github.com/GabeCordo/etl/core"
)

func main() {

	c := core.NewCore()

	//Vec := Vector{}

	//config := cluster.DefaultConfig
	//config.Identifier = "Vec"
	//c.Cluster("Vec", Vec, config)

	c.Run()
}
