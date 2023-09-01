package main

import (
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/mango/controllers"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
func main() {

	cli := commandline.NewCommandLine()

	ic := cli.AddCommand("init", controllers.InitCommand{})
	ic.SetCategory("utils").SetDescription("initialize the global files required to start the core")

	dc := cli.AddCommand("doctor", controllers.DoctorCommand{})
	dc.SetCategory("utils").SetDescription("verify the integrity of the global files on the local system")

	lc := cli.AddCommand("logs", controllers.LogController{})
	lc.SetCategory("data").SetDescription(
		"used to view log files stored on the local system" +
			"\n\t\t[path] specify the name of the log file to output" +
			"\n\t\t[normal|warning|fatal] specify the priority of logs outputted " +
			"\n\t\t\t(can help if you want to find fatal errors)")

	stc := cli.AddCommand("stats", controllers.StatisticsController{})
	stc.SetCategory("data").SetDescription("used to view the statistic files stored on the system")

	sc := cli.AddCommand("start", controllers.StartCommand{})
	sc.SetCategory("utils").SetDescription("start the core on the local system")

	cli.Run()
}
