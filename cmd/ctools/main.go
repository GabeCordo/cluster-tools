package main

import (
	"github.com/GabeCordo/cluster-tools/cmd/ctools/controllers"
	"github.com/GabeCordo/commandline"
)

func main() {

	cli := commandline.NewCommandLine()

	// util controllers

	ic := cli.AddCommand("example", controllers.ExampleController{})
	ic.SetCategory("utils").SetDescription("an example processor node to test the gateway")

	rc := cli.AddCommand("run", controllers.RunController{})
	rc.SetCategory("utils").SetDescription("run a pipeline on the gateway")

	pc := cli.AddCommand("pipeline", controllers.PipelineController{})
	pc.SetCategory("utils").SetDescription("register a pipeline on the gateway")

	// configuration controllers

	nc := cli.AddCommand("namespace", controllers.NamespaceController{})
	nc.SetCategory("config").SetDescription("switch the namespace used by ctools")

	gc := cli.AddCommand("gate", controllers.GatewayController{})
	gc.SetCategory("config").SetDescription("switch the gateway used by ctools")

	// gateway controllers

	//cc := cli.AddCommand("connect", controllers.ConnectController{})
	//cc.SetCategory("gateway").SetDescription("connect the processor to a gateway")
	//
	//dc := cli.AddCommand("disconnect", controllers.DisconnectController{})
	//dc.SetCategory("gateway").SetDescription("disconnect the processor from a gateway")

	cli.Run()
}
