package controllers

import "github.com/GabeCordo/commandline"

type RunCommand struct {
}

func (controller RunCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	return commandline.Terminate
}
