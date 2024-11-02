package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/Sentmint/PipelineOps/cmd/ctools/local"
)

type NamespaceController struct {
}

func (controller NamespaceController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cli.Flag(commandline.Switch) {
		controller.switchFlag(cli)
	} else if cli.Flag(commandline.Show) {
		controller.showFlag(cli)
	}

	return commandline.Terminate
}

func (controller NamespaceController) switchFlag(cli *commandline.CommandLine) {

	namespace := cli.NextArg()
	if namespace == commandline.FinalArg {
		fmt.Println("no namspace specified")
		return
	}

	local.SwitchNamespace(namespace)
}

func (controller NamespaceController) showFlag(cli *commandline.CommandLine) {
	namespace := local.GetNamespace()
	fmt.Println(namespace)
}
