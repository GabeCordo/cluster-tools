package main

import (
	"github.com/FortifiedCode/commandline"
	"github.com/FortifiedCode/flock/cmd/flock/controllers"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
func main() {

	cli := commandline.NewCommandLine()

	ic := cli.AddCommand("init", controllers.InitCommand{})
	ic.SetCategory("utils").SetDescription("initialize the global files required to start the flock")

	dc := cli.AddCommand("doctor", controllers.DoctorCommand{})
	dc.SetCategory("utils").SetDescription("verify the integrity of the global files on the log system")

	lc := cli.AddCommand("logs", controllers.LogController{})
	lc.SetCategory("data").SetDescription(
		"used to view log files stored on the log system" +
			"\n\t\t[path] specify the name of the log file" +
			"\n\t\t[normal|warning|fatal] specify the log priority " +
			"\n\t\t\t(ex. find fatal errors)")

	stc := cli.AddCommand("stats", controllers.StatisticsController{})
	stc.SetCategory("data").SetDescription("used to view the statistic files stored on the system")

	sc := cli.AddCommand("start", controllers.StartCommand{})
	sc.SetCategory("utils").SetDescription("start the flock on the log system")

	rc := cli.AddCommand("repl", controllers.ReplController{})
	rc.SetCategory("utils").SetDescription("enable or disable the repl when running mango start")

	shc := cli.AddCommand("scheduler", controllers.ScheduleController{})
	shc.SetCategory("utils").SetDescription("create or delete schedules for when clusters should be provisioned")

	cc := cli.AddCommand("config", controllers.ConfigCommand{})
	cc.SetCategory("utils").SetDescription("update the global flock configuration")

	// util controllers

	ec := cli.AddCommand("example", controllers.ExampleController{})
	ec.SetCategory("utils").SetDescription("an example processor node to test the gateway")

	runc := cli.AddCommand("run", controllers.RunController{})
	runc.SetCategory("utils").SetDescription("run a pipeline on the gateway")

	pc := cli.AddCommand("pipeline", controllers.PipelineController{})
	pc.SetCategory("utils").SetDescription("register a pipeline on the gateway")

	schc := cli.AddCommand("schedule", controllers.ScheduleController{})
	schc.SetCategory("utils").SetDescription("schedule jobs on the gateway")

	statc := cli.AddCommand("jobs", controllers.StatsController{})
	statc.SetCategory("utils").SetDescription("view stats associated with jobs on flock")

	// configuration controllers

	nc := cli.AddCommand("namespace", controllers.NamespaceController{})
	nc.SetCategory("config").SetDescription("switch the namespace used by pops")

	gc := cli.AddCommand("gate", controllers.GatewayController{})
	gc.SetCategory("config").SetDescription("switch the gateway used by pops")

	cli.Run()
}
