package main

import (
	"github.com/GabeCordo/commandline"
	"github.com/Sentmint/PipelineOps/cmd/pops/controllers"
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

	sc := cli.AddCommand("schedule", controllers.ScheduleController{})
	sc.SetCategory("utils").SetDescription("schedule jobs on the gateway")

	stc := cli.AddCommand("stats", controllers.StatsController{})
	stc.SetCategory("utils").SetDescription("view stats associated with jobs on the gateway")

	// configuration controllers

	nc := cli.AddCommand("namespace", controllers.NamespaceController{})
	nc.SetCategory("config").SetDescription("switch the namespace used by pops")

	gc := cli.AddCommand("gate", controllers.GatewayController{})
	gc.SetCategory("config").SetDescription("switch the gateway used by pops")

	cli.Run()
}
