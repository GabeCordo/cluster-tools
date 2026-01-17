package controllers

import (
	"fmt"

	"github.com/FortifiedCode/commandline"
	"github.com/FortifiedCode/flock/cmd/flock/local"
)

type OriginController struct {
}

func (controller OriginController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Show) {
		controller.showFlag(cli)
	} else if cli.Flag(commandline.Update) {
		controller.switchFlag(cli)
	} else {
		fmt.Println("[!] origin expects a 'show' or 'update' flag passed to it.")
	}

	return commandline.Terminate
}

func (controller OriginController) showFlag(cli *commandline.CommandLine) {

	core := local.GetCore()
	fmt.Println(core)
}

func (controller OriginController) switchFlag(cli *commandline.CommandLine) {

	newGateway := cli.NextArg()
	if newGateway == commandline.FinalArg {
		fmt.Println("you need to specify a new gateway")
	}

	if newGateway[len(newGateway)-1:] == "/" {
		newGateway = newGateway[:len(newGateway)-1]
	}

	local.SwitchCore(newGateway)
}
