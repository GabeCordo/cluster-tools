package controllers

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	etlCore "github.com/GabeCordo/etl/core"
	"github.com/GabeCordo/fack"
)

type PermissionCommand struct {
	PublicName string
}

func (pc PermissionCommand) Name() string {
	return pc.PublicName
}

func (pc PermissionCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	if cl.Flags.Add {
		pc.AddPermissionToEndpoint(cl)
	}
	return true // complete
}

func (pc PermissionCommand) AddPermissionToEndpoint(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	globalOrLocalFlag := cl.NextArg()
	if globalOrLocalFlag == commandline.FinalArg {
		fmt.Println("missing local or global flag")
		return true
	}

	if !((globalOrLocalFlag == "local") || (globalOrLocalFlag == "global")) {
		fmt.Println("you must specify whether the permission is global or local")
		return true
	}

	endpointIdentity := cl.NextArg()
	if endpointIdentity == commandline.FinalArg {
		fmt.Println("missing endpoint identity")
		return true
	}

	var localPermissionRoute string
	if globalOrLocalFlag == "local" {
		localPermissionRoute = cl.NextArg()
		if localPermissionRoute == commandline.FinalArg {
			fmt.Printf("you requested to change a local permission on the endpoint '%s'", endpointIdentity)
			return true
		}
	}

	method := cl.NextArg()
	if method == commandline.FinalArg {
		fmt.Println("missing http method")
		return true
	}

	if !fack.IsValidHTTPMethod(method) {
		fmt.Println("'" + method + "' is not a valid HTTP method")
		return true
	}

	enableMethod := cl.NextArg()
	if enableMethod == commandline.FinalArg {
		fmt.Println("missing directive to enable or disable HTTP method")
		return true
	}

	if !((enableMethod == "enable") || (enableMethod == "disable")) {
		fmt.Println("directive for the HTTP method must be 'enable' or 'disable'")
		return true
	}
	httpMethod := fack.HTTPMethodFromString(method)

	configPath := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory).File("config.etl.json")
	if configPath.DoesNotExist() {
		fmt.Println("no etl project found")
		return true
	}

	config := etlCore.NewConfig("tmp")
	etlCore.JSONToETLConfig(config, configPath.ToString())

	if _, found := config.Auth.Trusted[endpointIdentity]; !found {
		fmt.Println("no endpoint with that identity exists")
		return true
	}

	endpoint := config.Auth.Trusted[endpointIdentity]
	if globalOrLocalFlag == "global" {
		if enableMethod == "enable" {
			endpoint.GlobalPermissions.Enable(httpMethod)
		} else {
			endpoint.GlobalPermissions.Disable(httpMethod)
		}
	} else {
		// if the permission endpoint does not exist, it must be added
		if _, err := endpoint.LocalPermission(localPermissionRoute); err != nil {
			endpoint.AddLocalPermission(localPermissionRoute, fack.NewPermission().NoAccess())
		}
		localPermission := endpoint.LocalPermissions[localPermissionRoute]
		if enableMethod == "enable" {
			localPermission.Enable(httpMethod)
		} else {
			localPermission.Disable(httpMethod)
		}
	}

	config.ToJson(configPath.ToString())

	return true
}
