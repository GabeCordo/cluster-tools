package controllers

import (
	"fmt"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/api"

	"github.com/FortifiedCode/commandline"
)

type DisconnectController struct {
}

func (controller DisconnectController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	gatewayHost := cli.NextArg()
	if gatewayHost == commandline.FinalArg {
		fmt.Println("[!] missing first argument, expected gateway host")
		return commandline.Terminate
	}

	err := api.Disconnect(gatewayHost)
	if err != nil {
		fmt.Println("[!] error disconnecting from the gateway")
	} else {
		fmt.Println("[-] disconnected from the gateway")
	}

	return commandline.Terminate
}
