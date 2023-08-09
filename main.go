package main

import (
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/controllers"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
func main() {

	cli := commandline.NewCommandLine()

	cli.AddCommand("doctor", controllers.DoctorCommand{}).SetCategory("utils")
	cli.AddCommand("start", controllers.StartCommand{}).SetCategory("utils")
	cli.AddCommand("init", controllers.InitCommand{}).SetCategory("utils")

	cli.AddCommand("module", controllers.ModuleCommand{}).SetCategory("modules")

	cli.Run()
}
