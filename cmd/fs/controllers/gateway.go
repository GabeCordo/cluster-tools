package controllers

import (
	"fmt"

	"github.com/FortifiedCode/commandline"
	"github.com/GabeCordo/FunctionScheduler/cmd/FunctionScheduler/local"
)

type GatewayController struct {
}

func (controller GatewayController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Show) {
		controller.showFlag(cli)
	} else {
		controller.switchFlag(cli)
	}

	return commandline.Terminate
}

func (controller GatewayController) showFlag(cli *commandline.CommandLine) {

	core := local.GetCore()
	fmt.Println(core)
}

func (controller GatewayController) switchFlag(cli *commandline.CommandLine) {

	newGateway := cli.NextArg()
	if newGateway == commandline.FinalArg {
		fmt.Println("you need to specify a new gateway")
	}

	if newGateway[len(newGateway)-1:] == "/" {
		newGateway = newGateway[:len(newGateway)-1]
	}

	local.SwitchCore(newGateway)
}
