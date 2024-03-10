package http_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/core/components/processor"
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"net/http"
	"net/url"
	"strconv"
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

	processors, success := common.GetProcessors(thread.C5, thread.ProcessorResponseTable, thread.config.Timeout)

	response := interfaces.HTTPResponse{Success: success}

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

	success, modules := common.GetModules(thread.C5, thread.ProcessorResponseTable, thread.config.Timeout)

	response := interfaces.HTTPResponse{Success: success}
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
		success, err = common.MountModule(thread.C5, thread.ProcessorResponseTable, request.ModuleName, thread.config.Timeout)
	} else {
		success, err = common.UnmountModule(thread.C5, thread.ProcessorResponseTable, request.ModuleName, thread.config.Timeout)
	}

	response := interfaces.HTTPResponse{Success: success}

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

	clusterList, success := common.GetClusters(thread.C5, thread.ProcessorResponseTable, moduleName[0], thread.config.Timeout)
	if !success {
		w.WriteHeader(http.StatusNotFound)
	}

	response := interfaces.HTTPResponse{Success: success}

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

	response := interfaces.HTTPResponse{}

	if request.Mounted {
		response.Success = common.MountCluster(thread.C5, thread.ProcessorResponseTable, request.Module, request.Cluster, thread.config.Timeout)
	} else {
		response.Success = common.UnmountCluster(thread.C5, thread.ProcessorResponseTable, request.Module, request.Cluster, thread.config.Timeout)
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
	Config     string            `json:"config"`
	Supervisor uint64            `json:"id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type SupervisorProvisionJSONResponse struct {
	Cluster    string `json:"cluster,omitempty"`
	Supervisor uint64 `json:"id,omitempty"`
}

func (thread *Thread) supervisorCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		thread.getSupervisorCallback(w, r)
	} else if r.Method == "POST" {
		thread.postSupervisorCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getSupervisorCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	idStr, idFound := urlMapping["id"]
	if !idFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &interfaces.HTTPResponse{Success: true}

	instance, err := common.GetSupervisor(thread.C5, thread.ProcessorResponseTable, thread.config.Timeout, id)
	if err != nil {
		response.Success = false
		response.Description = err.Error()
	} else {
		response.Data = instance
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) postSupervisorCallback(w http.ResponseWriter, r *http.Request) {

	var request SupervisorConfigJSONBody

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if supervisorId, err := common.CreateSupervisor(
		thread.C5,
		thread.ProcessorResponseTable,
		request.Module,
		request.Cluster,
		request.Config,
		request.Metadata,
		thread.config.Timeout,
	); err == nil {

		response := &SupervisorProvisionJSONResponse{Cluster: request.Cluster, Supervisor: supervisorId}
		bytes, _ := json.Marshal(response)
		if _, err := w.Write(bytes); err != nil {
			// TODO : support module is not mounted
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (thread *Thread) configCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	request := &interfaces.Config{}
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
			if config, found := common.GetConfigFromDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0], clusterName[0], thread.config.Timeout); found {
				bytes, _ := json.Marshal(config)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			if configs, found := common.GetConfigsFromDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0], thread.config.Timeout); found {
				bytes, _ := json.Marshal(configs)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}

	} else if r.Method == "POST" {

		err := common.StoreConfigInDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0], *request, thread.config.Timeout)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
		}

	} else if r.Method == "PUT" {
		isOk := common.ReplaceConfigInDatabase(thread.C1, thread.DatabaseResponseTable, moduleName[0], *request, thread.config.Timeout)
		if !isOk {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Method == "DELETE" {

		clusterName, foundClusterName := urlMapping["config"]

		if foundClusterName {
			if isOk := common.DeleteConfigInDatabase(thread.C1, thread.DatabaseResponseTable,
				moduleName[0], clusterName[0], thread.config.Timeout); !isOk {
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
			statistics, found := common.FindStatistics(thread.C1, thread.DatabaseResponseTable, moduleName[0], clusterName[0], thread.config.Timeout)
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

func (thread *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		thread.getDebugCallback(w, r)
	} else if r.Method == "POST" {
		thread.postDebugCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getDebugCallback(w http.ResponseWriter, r *http.Request) {

	thread.logger.Printf("ping from %s\n", r.RemoteAddr)
	response := interfaces.HTTPResponse{Success: true, Description: "bonjour"}
	b, _ := json.Marshal(response)
	w.Write(b)
}

type DebugJSONBody struct {
	Action string `json:"action"`
}

type DebugJSONResponse struct {
	Duration time.Duration `json:"time-elapsed"`
	Success  bool          `json:"success"`
}

func (thread *Thread) postDebugCallback(w http.ResponseWriter, r *http.Request) {

	var request DebugJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("missing body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := interfaces.HTTPResponse{Success: true}

	if request.Action == "shutdown" {
		err = common.ShutdownCore(thread.Interrupt)
		if err != nil {
			response.Description = err.Error()
		}
	} else if request.Action == "ping" {
		startTime := time.Now()
		err = common.PingNodeChannels(thread.C1, thread.DatabaseResponseTable,
			thread.C5, thread.ProcessorResponseTable, thread.config.Timeout)
		if err == nil {
			duration := time.Now().Sub(startTime)
			response.Data = duration
		} else {
			response.Description = err.Error()
		}
	}

	response.Success = err == nil
	b, _ := json.Marshal(response)
	w.Write(b)
}
