package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/config"
	"github.com/GabeCordo/cluster-tools/internal/database/job"
	"github.com/GabeCordo/cluster-tools/internal/processor"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/cluster-tools/internal/thread/rest"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// TODO : add comments to the else conditions where the processor may support

func (t *Thread) processorCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		/* show the operator all the processors attached to the cluster-tools */
		t.getProcessorCallback(w, r)
	} else {
		/* the client does not support any other methods on the processor */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getProcessorCallback(w http.ResponseWriter, r *http.Request) {

	processors, success := thread.GetProcessors(
		thread.Mandatory{
			t.C5,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
	)

	response := rest.Response{Success: success}

	if success {
		response.Data = processors
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (t *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		/* show the operator all the modules registered to the cluster-tools */
		t.getModuleCallback(w, r)
	} else if r.Method == "PUT" {
		/* the operator shall be allowed to mount and unmount modules */
		t.putModuleCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getModuleCallback(w http.ResponseWriter, r *http.Request) {

	success, modules := thread.GetModules(
		thread.Mandatory{
			t.C5,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
	)

	response := rest.Response{Success: success}
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

func (t *Thread) putModuleCallback(w http.ResponseWriter, r *http.Request) {

	request := &ModuleBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	/* database the success of the request in this address */
	var success bool = false

	mandatory := thread.Mandatory{t.C5, t.ProcessorResponseTable, t.config.Timeout}

	if request.Mounted {
		success, err = thread.MountModule(mandatory, request.ModuleName)
	} else {
		success, err = thread.UnmountModule(mandatory, request.ModuleName)
	}

	response := rest.Response{Success: success}

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

func (t *Thread) clusterCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		/* the operator shall see clusters registered to the cluster-tools */
		t.getClusterCallback(w, r)
	} else if r.Method == "PUT" {
		/* the operator shall mount clusters in the cluster-tools */
		/* the operator shall unmount clusters in the cluster-tools */
		t.putClusterCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getClusterCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	moduleName, foundModuleName := urlMapping["module"]

	if !foundModuleName {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clusterList, success := thread.GetClusters(
		thread.Mandatory{
			t.C5,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
		moduleName[0],
	)
	if !success {
		w.WriteHeader(http.StatusNotFound)
	}

	response := rest.Response{Success: success}

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

func (t *Thread) putClusterCallback(w http.ResponseWriter, r *http.Request) {

	request := &ClusterConfigJSONBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := rest.Response{}

	mandatory := thread.Mandatory{t.C5, t.ProcessorResponseTable, t.config.Timeout}

	if request.Mounted {
		response.Success = thread.MountCluster(mandatory, request.Module, request.Cluster)
	} else {
		response.Success = thread.UnmountCluster(mandatory, request.Module, request.Cluster)
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

func (t *Thread) supervisorCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t.getSupervisorCallback(w, r)
	} else if r.Method == "POST" {
		t.postSupervisorCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getSupervisorCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	module := ""
	if moduleStr, found := urlMapping["module"]; found {
		module = moduleStr[0]
	}

	cluster := ""
	if clusterStr, found := urlMapping["cluster"]; found {
		cluster = clusterStr[0]
	}

	var id string
	if idStr, found := urlMapping["id"]; found {
		if _, err := strconv.ParseUint(idStr[0], 10, 64); err != nil {
			id = idStr[0]
		} else {
			id = "0"
		}
	} else {
		id = "0"
	}

	if (module == "") && (cluster == "") && (id == "0") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &rest.Response{Success: true}

	mandatory := thread.Mandatory{t.C5, t.ProcessorResponseTable, t.config.Timeout}
	filter := database.Filter{Module: module, Cluster: cluster, Identifier: id}

	instance, err := thread.GetSupervisor(mandatory, filter)
	if err != nil {
		response.Success = false
		response.Description = err.Error()
	} else {
		response.Data = instance
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (t *Thread) postSupervisorCallback(w http.ResponseWriter, r *http.Request) {

	var request SupervisorConfigJSONBody

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if supervisorId, err := thread.CreateSupervisor(
		thread.Mandatory{
			t.C5,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
		request.Module,
		request.Cluster,
		request.Config,
		request.Metadata,
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

func (t *Thread) configCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	request := &config.Config{}
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

	mandatory := thread.Mandatory{t.C1, t.DatabaseResponseTable, t.config.Timeout}

	if r.Method == "GET" {

		clusterName, foundClusterName := urlMapping["config"]

		if foundClusterName {
			if cfg, found := thread.GetConfigFromDatabase(mandatory, moduleName[0], clusterName[0]); found {
				bytes, _ := json.Marshal(cfg)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			if configs, found := thread.GetConfigsFromDatabase(mandatory, moduleName[0]); found {
				bytes, _ := json.Marshal(configs)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}

	} else if r.Method == "POST" {

		err := thread.StoreConfigInDatabase(mandatory, moduleName[0], *request)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
		}

	} else if r.Method == "PUT" {
		isOk := thread.ReplaceConfigInDatabase(mandatory, moduleName[0], *request)
		if !isOk {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Method == "DELETE" {

		configName, foundConfigName := urlMapping["config"]

		if foundConfigName {
			if isOk := thread.DeleteConfigInDatabase(mandatory, moduleName[0], configName[0]); !isOk {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (t *Thread) statisticCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	mandatory := thread.Mandatory{t.C1, t.DatabaseResponseTable, t.config.Timeout}

	if r.Method == "GET" {

		moduleName, moduleNameFound := urlMapping["module"]
		clusterName, clusterNameFound := urlMapping["cluster"]

		if moduleNameFound && clusterNameFound {
			statistics, found := thread.FindStatistics(mandatory, moduleName[0], clusterName[0])
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

func (t *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t.getDebugCallback(w, r)
	} else if r.Method == "POST" {
		t.postDebugCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getDebugCallback(w http.ResponseWriter, r *http.Request) {

	t.logger.Printf("ping from %s\n", r.RemoteAddr)
	response := rest.Response{Success: true, Description: "bonjour"}
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

func (t *Thread) postDebugCallback(w http.ResponseWriter, r *http.Request) {

	var request DebugJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("missing body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := rest.Response{Success: true}

	if request.Action == "shutdown" {
		err = thread.ShutdownCore(t.Interrupt)
		if err != nil {
			response.Description = err.Error()
		}
	}

	response.Success = err == nil
	b, _ := json.Marshal(response)
	w.Write(b)
}

func (t *Thread) jobCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getJobCallback(w, r)
	} else if r.Method == http.MethodPost {
		t.postJobCallback(w, r)
	} else if r.Method == http.MethodDelete {
		t.deleteJobCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getJobCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	identifier := ""
	if tmp, found := urlMapping["id"]; found {
		identifier = tmp[0]
	}
	module := ""
	if tmp, found := urlMapping["module"]; found {
		module = tmp[0]
	}
	cluster := ""
	if tmp, found := urlMapping["cluster"]; found {
		cluster = tmp[0]
	}
	minutes := 0
	if tmp, found := urlMapping["minutes"]; found {
		if i, err := strconv.Atoi(tmp[0]); err != nil {
			minutes = i
		}
	}

	filter := &database.Filter{
		Identifier: identifier,
		Module:     module,
		Cluster:    cluster,
		Interval: database.Interval{
			Minute: minutes,
		}}

	var err error

	response := rest.Response{}
	response.Data, err = thread.GetJobs(thread.Mandatory{t.C20, t.SchedulerResponseTable, t.config.Timeout}, filter)

	response.Success = err == nil
	if err != nil {
		response.Description = err.Error()
	}

	if b, err := json.Marshal(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(b)
	}
}

func (t *Thread) postJobCallback(w http.ResponseWriter, r *http.Request) {

	var job job.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		fmt.Println("missing job passed to body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := rest.Response{}
	if err := thread.CreateJob(thread.Mandatory{t.C20, t.SchedulerResponseTable, t.config.Timeout}, &job); err != nil {
		response.Success = false
		response.Data = err.Error()
	} else {
		response.Success = true
	}

	if b, err := json.Marshal(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(b)
	}
}

func (t *Thread) deleteJobCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	identifier := ""
	if tmp, found := urlMapping["id"]; found {
		identifier = tmp[0]
	}
	module := ""
	if tmp, found := urlMapping["module"]; found {
		module = tmp[0]
	}
	cluster := ""
	if tmp, found := urlMapping["cluster"]; found {
		cluster = tmp[0]
	}
	minutes := 0
	if tmp, found := urlMapping["minutes"]; found {
		if i, err := strconv.Atoi(tmp[0]); err != nil {
			minutes = i
		}
	}

	filter := &database.Filter{
		Identifier: identifier,
		Module:     module,
		Cluster:    cluster,
		Interval: database.Interval{
			Minute: minutes,
		}}

	var err error

	response := rest.Response{}
	err = thread.DeleteJob(
		thread.Mandatory{
			t.C20,
			t.SchedulerResponseTable,
			t.config.Timeout},
		filter,
	)

	response.Success = err == nil
	if err != nil {
		response.Description = err.Error()
	}

	if b, err := json.Marshal(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(b)
	}
}

func (t *Thread) jobQueueCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getJobQueueCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getJobQueueCallback(w http.ResponseWriter, r *http.Request) {

	response := rest.Response{}

	var err error
	response.Data, err = thread.JobQueue(thread.Mandatory{t.C20, t.SchedulerResponseTable, t.config.Timeout})

	if err != nil {
		response.Description = err.Error()
	}
	response.Success = err == nil

	if b, err := json.Marshal(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(b)
	}
}
