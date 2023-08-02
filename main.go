package main

import (
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/controllers"
)

func main() {

	cli := commandline.NewCommandLine()

	cli.AddCommand("doctor", controllers.DoctorCommand{}).SetCategory("utils")
	cli.AddCommand("start", controllers.StartCommand{}).SetCategory("utils")
	cli.AddCommand("init", controllers.InitCommand{}).SetCategory("utils")

	cli.AddCommand("module", controllers.ModuleCommand{}).SetCategory("modules")

	cli.Run()
}
