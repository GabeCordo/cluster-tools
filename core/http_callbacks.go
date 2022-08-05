package core

import (
	"ETLFramework/net"
)

func (http Http) ClustersFunction(request *net.Request, response *net.Response) {
	supervisorRequest := SupervisorRequest{Provision, request.Function, request.Param}
	http.C5 <- supervisorRequest

	response.AddStatus(200, net.Success)
}

func (http Http) StatisticsFunction(request *net.Request, response *net.Response) {
	// TODO - not implemented
}

func (http Http) DebugFunction(request *net.Request, response *net.Response) {
	if request.Function == "shutdown" {
		http.Interrupt <- Shutdown
	} else if request.Function == "endpoints" {
		auth := GetAuthInstance()

		if len(request.Param) == 1 {
			if endpoint, found := auth.Trusted[request.Param[0]]; found {
				response.AddPair("localPermission", endpoint.LocalPermissions)
				response.AddPair("globalPermission", endpoint.GlobalPermissions)
			} else {
				response.AddStatus(400, net.BadArgument)
				return
			}
		} else {
			var endpoints []string
			for key, _ := range auth.Trusted {
				endpoints = append(endpoints, key)
			}

			response.AddPair("endpoints", endpoints)
		}

		response.AddStatus(200, net.Success)
	}
}
