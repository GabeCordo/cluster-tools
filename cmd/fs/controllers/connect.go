package controllers

import (
	"fmt"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"

	"github.com/GabeCordo/Commandline"
)

type ConnectController struct {
}

func (controller ConnectController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	gatewayHost := cli.NextArg()
	if gatewayHost == commandline.FinalArg {
		fmt.Println("[!] missing first argument, expected gateway host")
		return commandline.Terminate
	}

	err := api.Connect(gatewayHost)
	if err != nil {
		fmt.Println(err)
		fmt.Println("[!] error connecting to gateway")
	} else {
		fmt.Println("[+] connected to gateway")
	}

	return commandline.Terminate
}
