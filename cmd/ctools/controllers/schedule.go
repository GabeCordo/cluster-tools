package controllers

import "github.com/GabeCordo/commandline"

type ScheduleController struct {
}

func (controller ScheduleController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Show) {
		controller.showFlag(cli)
	} else if cli.Flag(commandline.Add) {
		controller.addFlag(cli)
	} else if cli.Flag(commandline.Delete) {
		controller.deleteFlag(cli)
	}

	return commandline.Terminate
}

func (controller ScheduleController) showFlag(cli *commandline.CommandLine) {

}

func (controller ScheduleController) addFlag(cli *commandline.CommandLine) {

}

func (controller ScheduleController) deleteFlag(cli *commandline.CommandLine) {

}
