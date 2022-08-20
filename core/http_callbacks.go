package core

import (
	"ETLFramework/net"
	"math/rand"
)

func (http *HttpThread) ClustersFunction(request *net.Request, response *net.Response) {
	supervisorRequest := ProvisionerRequest{Provision, rand.Uint32(), request.Function, request.Param}
	http.C5 <- supervisorRequest

	response.AddStatus(200)
}

func (http *HttpThread) StatisticsFunction(request *net.Request, response *net.Response) {
	req := DatabaseRequest{Action: Fetch, Cluster: request.Function}

	if value, ok := http.Send(Database, req); ok {
		rsp := (value).(DatabaseResponse)

		// check to see if no records have ever been created
		if !rsp.Success {
			response.AddStatus(200, "no cluster records exist")
			return
		}
		response.AddPair("value", rsp.Data)
	}

	response.AddStatus(200)
}

func (http *HttpThread) DebugFunction(request *net.Request, response *net.Response) {
	if request.Function == "shutdown" {
		http.Interrupt <- Shutdown
	} else if request.Function == "lookup" {
		provisioner := GetProvisionerInstance()
		if len(request.Param) == 1 {
			if _, found := provisioner.Functions[request.Param[0]]; found {
				// the cluster identifier exists on the node and can be called
				response.AddStatus(200)
			} else {
				// the cluster identifier does NOT exist, return "not found"
				response.AddStatus(404)
			}
		} else {
			response.AddStatus(400, "missing cluster identifier")
		}
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

		response.AddStatus(200)
	} else {
		// output system information
		config := GetConfigInstance()
		response.AddPair("name", config.Name)
		response.AddPair("version", config.Version)

		response.AddStatus(200)
	}
}
