package main

import (
	"github.com/GabeCordo/commandline"
	"github.com/Sentmint/PipelineOps/cmd/ctgate/controllers"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
func main() {

	cli := commandline.NewCommandLine()

	ic := cli.AddCommand("init", controllers.InitCommand{})
	ic.SetCategory("utils").SetDescription("initialize the global files required to start the ctgate")

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
	sc.SetCategory("utils").SetDescription("start the ctgate on the log system")

	rc := cli.AddCommand("repl", controllers.ReplController{})
	rc.SetCategory("utils").SetDescription("enable or disable the repl when running mango start")

	shc := cli.AddCommand("scheduler", controllers.ScheduleController{})
	shc.SetCategory("utils").SetDescription("create or delete schedules for when clusters should be provisioned")

	cc := cli.AddCommand("pipeline", controllers.ConfigCommand{})
	cc.SetCategory("utils").SetDescription("update the global ctgate configuration")

	cli.Run()
}
