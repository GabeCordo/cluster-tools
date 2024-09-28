package main

import (
	"github.com/GabeCordo/cluster-tools/cmd/ctools/controllers"
	"github.com/GabeCordo/commandline"
)

func main() {

	cli := commandline.NewCommandLine()

	// other controllers

	ic := cli.AddCommand("example", controllers.ExampleController{})
	ic.SetCategory("other").SetDescription("an example processor node to test the gateway")

	// standalone (or) gateway controllers

	rc := cli.AddCommand("run", controllers.RunCommand{})
	rc.SetCategory("utils").SetDescription("run a pipeline on the processor")

	// gateway controllers

	cc := cli.AddCommand("connect", controllers.ConnectController{})
	cc.SetCategory("gateway").SetDescription("connect the processor to a gateway")

	dc := cli.AddCommand("disconnect", controllers.DisconnectController{})
	dc.SetCategory("gateway").SetDescription("disconnect the processor from a gateway")

	cli.Run()
}
