package core

import (
	"github.com/GabeCordo/fack"
	"github.com/GabeCordo/fack/rpc"
	"math/rand"
	"strconv"
)

// TODO - fix rpc to request conversion

func (http *HttpThread) ClustersFunction(request fack.Request, response fack.Response) {
	rpcRequest := request.(*rpc.Request)

	if len(rpcRequest.Param) > 1 {
		response.SetStatus(400)
		return
	}

	provisionerThreadRequest := ProvisionerRequest{Nonce: rand.Uint32(), Cluster: rpcRequest.Param[0], Parameters: rpcRequest.Param}
	if rpcRequest.Function == "provision" {
		provisionerThreadRequest.Action = Provision
	} else if rpcRequest.Function == "mount" {
		provisionerThreadRequest.Action = Mount
	} else if rpcRequest.Function == "unmount" {
		provisionerThreadRequest.Action = UnMount
	}

	http.C5 <- provisionerThreadRequest
	response.SetStatus(200)
}

func (http *HttpThread) StatisticsFunction(request fack.Request, response fack.Response) {
	rpcRequest := request.(*rpc.Request)

	req := DatabaseRequest{Action: Fetch, Cluster: rpcRequest.Function}

	if value, ok := http.Send(Database, req); ok {
		rsp := (value).(DatabaseResponse)

		// check to see if no records have ever been created
		if !rsp.Success {
			response.SetStatus(200).SetDescription("no cluster records exist")
			return
		}
		response.Pair("value", rsp.Data)
	}

	response.SetStatus(200)
}

func (http *HttpThread) DataFunction(request fack.Request, response fack.Response) {
	rpcRequest := request.(*rpc.Request)

	statusCode := 200
	statusString := "no error"

	if rpcRequest.Function == "mounts" {
		provisionerInstance := GetProvisionerInstance()

		mounts := provisionerInstance.Mounts()
		for identifier, isMounted := range mounts {
			response.Pair(identifier, isMounted)
		}
	} else if rpcRequest.Function == "supervisor" {
		provisionerInstance := GetProvisionerInstance()

		if len(rpcRequest.Param) >= 1 {
			supervisorRequest := rpcRequest.Param[0]

			if supervisorRequest == "lookup" {
				if len(rpcRequest.Param) == 2 {
					clusterIdentifier := rpcRequest.Param[1]

					if _, found := provisionerInstance.RegisteredFunctions[clusterIdentifier]; found {
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
				if len(rpcRequest.Param) >= 2 {
					clusterIdentifier := rpcRequest.Param[1]

					registry, found := provisionerInstance.Registries[clusterIdentifier]
					if found {
						if len(rpcRequest.Param) == 3 {
							supervisorId := rpcRequest.Param[2]

							id, _ := strconv.ParseUint(supervisorId, 10, 32)
							supervisor, found := registry.GetSupervisor(id)
							if found {
								response.Pair("state", supervisor.State.String())
							} else {
								statusCode = 400
								statusString = "unknown supervisor id"
							}
						} else {
							for _, supervisor := range registry.Supervisors {
								id := strconv.FormatUint(supervisor.Id, 10)
								response.Pair(id, supervisor.State.String())
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
			} else {
				// display all relevant information about the supervisor
				if len(rpcRequest.Param) == 2 {
					clusterIdentifier := rpcRequest.Param[0]
					registry, ok := provisionerInstance.Registries[clusterIdentifier]
					if ok {
						supervisorIdStr := rpcRequest.Param[1]
						supervisorId, err := strconv.ParseUint(supervisorIdStr, 10, 64)
						if err == nil {
							supervisor, ok := registry.GetSupervisor(supervisorId)
							if ok {
								response.Pair("id", supervisor.Id)
								response.Pair("state", supervisor.State.String())
								response.Pair("num-e-routines", supervisor.Stats.NumProvisionedExtractRoutines)
								response.Pair("num-t-routines", supervisor.Stats.NumProvisionedTransformRoutes)
								response.Pair("num-l-routines", supervisor.Stats.NumProvisionedLoadRoutines)
								response.Pair("num-et-breaches", supervisor.Stats.NumEtThresholdBreaches)
								response.Pair("num-tl-breaches", supervisor.Stats.NumTlThresholdBreaches)
							} else {
								statusCode = 400
								statusString = rpc.BadArgument
							}
						} else {
							statusCode = 400
							statusString = rpc.BadArgument
						}
					} else {
						statusCode = 400
						statusString = rpc.BadArgument
					}
				} else if len(rpcRequest.Param) == 1 {
					clusterIdentifier := rpcRequest.Param[0]
					registry, ok := provisionerInstance.Registries[clusterIdentifier]
					if ok {
						output := make(map[uint64]map[string]any)
						for id, supervisor := range registry.Supervisors {
							record := make(map[string]any)

							record["id"] = supervisor.Id
							record["state"] = supervisor.State.String()
							record["num-e-routines"] = supervisor.Stats.NumProvisionedExtractRoutines
							record["num-t-routines"] = supervisor.Stats.NumProvisionedTransformRoutes
							record["num-l-routines"] = supervisor.Stats.NumProvisionedLoadRoutines
							record["num-et-breaches"] = supervisor.Stats.NumEtThresholdBreaches
							record["num-tl-breaches"] = supervisor.Stats.NumTlThresholdBreaches

							output[id] = record
						}
						response.Pair("supervisors", output)
					} else {
						statusCode = 400
						statusString = rpc.BadArgument
					}
				} else {
					statusCode = 400
					statusString = rpc.SyntaxMismatch
				}
			}
		}
	}

	response.SetStatus(statusCode).SetDescription(statusString)
}

func (http *HttpThread) DebugFunction(request fack.Request, response fack.Response) {
	rpcRequest := request.(*rpc.Request)

	statusString := "no error"
	statusCode := 200

	if rpcRequest.Function == "shutdown" {
		http.Interrupt <- Shutdown
	} else if rpcRequest.Function == "endpoints" {
		auth := GetAuthInstance()

		if len(rpcRequest.Param) == 1 {
			endpointIdentifier := rpcRequest.Param[0]
			if endpoint, found := auth.Trusted[endpointIdentifier]; found {
				response.Pair("localPermission", endpoint.LocalPermissions)
				response.Pair("globalPermission", endpoint.GlobalPermissions)
			} else {
				statusCode = 400
				statusString = rpc.BadArgument
			}
		} else {
			var endpoints []string
			for key, _ := range auth.Trusted {
				endpoints = append(endpoints, key)
			}
			response.Pair("endpoints", endpoints)
		}
	} else {
		// output system information
		config := GetConfigInstance()
		response.Pair("name", config.Name)
		response.Pair("version", config.Version)
	}

	response.SetStatus(statusCode).SetDescription(statusString)
}
