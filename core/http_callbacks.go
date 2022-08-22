package core

import (
	"ETLFramework/net"
	"math/rand"
	"strconv"
)

func (http *HttpThread) ClustersFunction(request *net.Request, response *net.Response) {
	if len(request.Param) > 1 {
		response.AddStatus(400)
		return
	}

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: request.Param[0], Parameters: request.Param}
	if request.Function == "provision" {
		provisionerThreadRequest.Action = Provision
	} else if request.Function == "mount" {
		provisionerThreadRequest.Action = Mount
	} else if request.Function == "unmount" {
		provisionerThreadRequest.Action = UnMount
	}

	http.C5 <- provisionerThreadRequest
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
	statusString := "no error"
	statusCode := 200

	if request.Function == "shutdown" {
		http.Interrupt <- Shutdown
	} else if request.Function == "supervisor" {
		provisioner := GetProvisionerInstance()

		if len(request.Param) >= 1 {
			supervisorRequest := request.Param[0]

			if supervisorRequest == "lookup" {
				if len(request.Param) == 2 {
					clusterIdentifier := request.Param[1]

					if _, found := provisioner.RegisteredFunctions[clusterIdentifier]; found {
						// the cluster identifier exists on the node and can be called
					} else {
						// the cluster identifier does NOT exist, return "not found"
						statusCode = 404
					}
				} else {
					statusCode = 400
					statusString = "missing cluster identifier"
				}
			} else if supervisorRequest == "state" {
				if len(request.Param) >= 2 {
					clusterIdentifier := request.Param[1]

					registry, found := provisioner.Registries[clusterIdentifier]
					if found {
						if len(request.Param) == 3 {
							supervisorId := request.Param[2]

							id, _ := strconv.ParseUint(supervisorId, 10, 32)
							supervisor, found := registry.GetSupervisor(id)
							if found {
								response.AddPair("state", supervisor.State.String())
							} else {
								statusCode = 400
								statusString = "unknown supervisor id"
							}
						} else {
							for _, supervisor := range registry.Supervisors {
								id := strconv.FormatUint(supervisor.Id, 10)
								response.AddPair(id, supervisor.State.String())
							}
						}
					} else {
						statusCode = 400
						statusString = "unknown cluster identifier"
					}
				} else {
					statusCode = 400
					statusString = "missing cluster identifier"
				}
			}
		}
	} else if request.Function == "endpoints" {
		auth := GetAuthInstance()

		if len(request.Param) == 1 {
			endpointIdentifier := request.Param[0]
			if endpoint, found := auth.Trusted[endpointIdentifier]; found {
				response.AddPair("localPermission", endpoint.LocalPermissions)
				response.AddPair("globalPermission", endpoint.GlobalPermissions)
			} else {
				statusCode = 400
				statusString = net.BadArgument
			}
		} else {
			var endpoints []string
			for key, _ := range auth.Trusted {
				endpoints = append(endpoints, key)
			}
			response.AddPair("endpoints", endpoints)
		}
	} else {
		// output system information
		config := GetConfigInstance()
		response.AddPair("name", config.Name)
		response.AddPair("version", config.Version)
	}

	response.AddStatus(statusCode, statusString)
}
