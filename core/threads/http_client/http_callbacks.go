package http_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/mango-core/core/components/processor"
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango/components/cluster"
	"github.com/GabeCordo/mango/core"
	"net/http"
	"net/url"
	"time"
)

// TODO : add comments to the else conditions where the http_processor may support

func (thread *Thread) processorCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		/* show the operator all the processors attached to the core */
		thread.getProcessorCallback(w, r)
	} else {
		/* the http_client does not support any other methods on the processor */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getProcessorCallback(w http.ResponseWriter, r *http.Request) {

	processors, success := common.GetProcessors(thread.C5, thread.ProcessorResponseTable)

	response := core.Response{Success: success}

	if success {
		response.Data = processors
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		/* show the operator all the modules registered to the core */
		thread.getModuleCallback(w, r)
	} else if r.Method == "PUT" {
		/* the operator shall be allowed to mount and unmount modules */
		thread.putModuleCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getModuleCallback(w http.ResponseWriter, r *http.Request) {

	success, modules := common.GetModules(thread.C5, thread.ProcessorResponseTable)

	response := core.Response{Success: success}
	if success {
		response.Data = modules
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

type ModuleBody struct {
	ModuleName string `json:"module"`
	Mounted    bool   `json:"mounted"`
}

func (thread *Thread) putModuleCallback(w http.ResponseWriter, r *http.Request) {

	request := &ModuleBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	/* store the success of the request in this address */
	var success bool = false

	if request.Mounted {
		success, err = common.MountModule(thread.C5, thread.ProcessorResponseTable, request.ModuleName)
	} else {
		success, err = common.UnmountModule(thread.C5, thread.ProcessorResponseTable, request.ModuleName)
	}

	response := core.Response{Success: success}

	if errors.Is(err, processor.ModuleDoesNotExist) {
		w.WriteHeader(http.StatusNotFound)
	} else if !success {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) clusterCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		/* the operator shall see clusters registered to the core */
		thread.getClusterCallback(w, r)
	} else if r.Method == "PUT" {
		/* the operator shall mount clusters in the core */
		/* the operator shall unmount clusters in the core */
		thread.putClusterCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getClusterCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	moduleName, foundModuleName := urlMapping["module"]

	if !foundModuleName {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clusterList, success := common.GetClusters(thread.C5, thread.ProcessorResponseTable, moduleName[0])
	if !success {
		w.WriteHeader(http.StatusNotFound)
	}

	response := core.Response{Success: success}

	if success {
		response.Data = clusterList
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

type ClusterConfigJSONBody struct {
	Module  string `json:"module"`
	Cluster string `json:"cluster"`
	Mounted bool   `json:"mounted"`
}

func (thread *Thread) putClusterCallback(w http.ResponseWriter, r *http.Request) {

	request := &ClusterConfigJSONBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := core.Response{}

	if request.Mounted {
		response.Success = common.MountCluster(thread.C5, thread.ProcessorResponseTable, request.Module, request.Cluster)
	} else {
		response.Success = common.UnmountCluster(thread.C5, thread.ProcessorResponseTable, request.Module, request.Cluster)
	}

	if !response.Success {
		w.WriteHeader(http.StatusNotFound)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

type SupervisorConfigJSONBody struct {
	Module     string            `json:"module"`
	Cluster    string            `json:"cluster"`
	Config     string            `json:"common"`
	Supervisor uint64            `json:"id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type SupervisorProvisionJSONResponse struct {
	Cluster    string `json:"cluster,omitempty"`
	Supervisor uint64 `json:"id,omitempty"`
}

func (thread *Thread) supervisorCallback(w http.ResponseWriter, r *http.Request) {

	// TODO : fix
	//urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	var request SupervisorConfigJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if (r.Method != "GET") && (err != nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//if r.Method == "GET" {
	//
	//	clusterName, foundClusterName := urlMapping["cluster"]
	//	moduleName, foundModuleName := urlMapping["module"]
	//	supervisorIdStr, foundSupervisorId := urlMapping["id"]
	//
	//	if !foundClusterName || !foundModuleName {
	//		w.WriteHeader(http.StatusBadRequest)
	//	} else {
	//		if foundSupervisorId {
	//			supervisorId, err := strconv.ParseUint(supervisorIdStr[0], 10, 64)
	//			if err != nil {
	//				w.WriteHeader(http.StatusBadRequest)
	//				return
	//			}
	//
	//			supervisorInst, success := common.GetSupervisor(thread.C5, thread.ProcessorResponseTable,
	//				moduleName[0], clusterName[0], supervisorId)
	//			if !success {
	//				w.WriteHeader(http.StatusBadRequest)
	//				return
	//			}
	//
	//			supervisorBytes, err := json.Marshal(supervisorInst)
	//			if err != nil {
	//				w.WriteHeader(http.StatusInternalServerError)
	//				return
	//			}
	//
	//			_, err = w.Write(supervisorBytes)
	//			if err != nil {
	//				w.WriteHeader(http.StatusInternalServerError)
	//			}
	//		} else {
	//			supervisors, success := common.GetSupervisors(thread.C5, thread.ProcessorResponseTable, moduleName[0], clusterName[0])
	//			if !success {
	//				w.WriteHeader(http.StatusBadRequest)
	//				return
	//			}
	//
	//			supervisorsBytes, err := json.Marshal(supervisors)
	//			if err != nil {
	//				w.WriteHeader(http.StatusInternalServerError)
	//				return
	//			}
	//
	//			_, err = w.Write(supervisorsBytes)
	//			if err != nil {
	//				w.WriteHeader(http.StatusInternalServerError)
	//			}
	//		}
	//	}
	if r.Method == "POST" {
		if supervisorId, err := common.CreateSupervisor(
			thread.C5,
			thread.ProcessorResponseTable,
			request.Module,
			request.Cluster,
			request.Config,
			request.Metadata,
		); err == nil {

			response := &SupervisorProvisionJSONResponse{Cluster: request.Cluster, Supervisor: supervisorId}
			bytes, _ := json.Marshal(response)
			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) configCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	request := &cluster.Config{}
	err := json.NewDecoder(r.Body).Decode(request)
	if (r.Method != "GET") && (r.Method != "DELETE") && (err != nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	/* the module always needs to be included */
	moduleName, foundModuleName := urlMapping["module"]
	if !foundModuleName {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {

		clusterName, foundClusterName := urlMapping["config"]

		if foundClusterName {
			if config, found := common.GetConfigFromDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0], clusterName[0]); found {
				bytes, _ := json.Marshal(config)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			if configs, found := common.GetConfigsFromDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0]); found {
				bytes, _ := json.Marshal(configs)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}

	} else if r.Method == "POST" {

		isOk := common.StoreConfigInDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0], *request)
		if !isOk {
			w.WriteHeader(http.StatusConflict)
		}

	} else if r.Method == "PUT" {
		isOk := common.ReplaceConfigInDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0], *request)
		if !isOk {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Method == "DELETE" {

		clusterName, foundClusterName := urlMapping["config"]

		if foundClusterName {
			if isOk := common.DeleteConfigInDatabase(thread.C1, thread.DatabaseResponseTable,
				moduleName[0], clusterName[0]); !isOk {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (thread *Thread) statisticCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	if r.Method == "GET" {

		moduleName, moduleNameFound := urlMapping["module"]
		clusterName, clusterNameFound := urlMapping["cluster"]

		if moduleNameFound && clusterNameFound {
			statistics, found := common.FindStatistics(thread.C1, thread.DatabaseResponseTable, moduleName[0], clusterName[0])
			if found {
				bytes, err := json.Marshal(statistics)
				if err == nil {
					if _, err = w.Write(bytes); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type DebugJSONBody struct {
	Action string `json:"action"`
}

type DebugJSONResponse struct {
	Duration time.Duration `json:"time-elapsed"`
	Success  bool          `json:"success"`
}

func (thread *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	var request DebugJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if (r.Method != "OPTIONS") && (r.Method != "GET") && err != nil {
		fmt.Println("missing body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "GET" {
		fmt.Println(r.RemoteAddr)
		// treat this as a probe to the http server
	} else if r.Method == "POST" {
		if request.Action == "shutdown" {
			common.ShutdownCore(thread.Interrupt)
			// TODO : fix
			//} else if request.Action == "ping" {
			//	startTime := time.Now()
			//	success := common.PingNodeChannels(thread.logger, thread.C1, thread.DatabaseResponseTable, thread.C5, thread.ProcessorResponseTable)
			//	response := DebugJSONResponse{Success: success, Duration: time.Now().Sub(startTime)}
			//	bytes, err := json.Marshal(response)
			//	if err == nil {
			//		if _, err := w.Write(bytes); err != nil {
			//			w.WriteHeader(http.StatusInternalServerError)
			//		}
			//	} else {
			//		w.WriteHeader(http.StatusInternalServerError)
			//	}
		} else if request.Action == "debug" {
			description := common.ToggleDebugMode(thread.logger)
			if _, err := w.Write([]byte(description)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST, GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) corsCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Add("Access-Control-Allow-Origin", "*")
	}
}