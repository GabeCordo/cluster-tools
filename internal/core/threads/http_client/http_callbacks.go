package http_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/components/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/components/scheduler"
	"github.com/GabeCordo/cluster-tools/internal/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
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

	processors, success := common.GetProcessors(
		common.ThreadMandatory{
			thread.C5,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
	)

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

	success, modules := common.GetModules(
		common.ThreadMandatory{
			thread.C5,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
	)

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

	mandatory := common.ThreadMandatory{thread.C5, thread.ProcessorResponseTable, thread.config.Timeout}

	if request.Mounted {
		success, err = common.MountModule(mandatory, request.ModuleName)
	} else {
		success, err = common.UnmountModule(mandatory, request.ModuleName)
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

	clusterList, success := common.GetClusters(
		common.ThreadMandatory{
			thread.C5,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
		moduleName[0],
	)
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

	mandatory := common.ThreadMandatory{thread.C5, thread.ProcessorResponseTable, thread.config.Timeout}

	if request.Mounted {
		response.Success = common.MountCluster(mandatory, request.Module, request.Cluster)
	} else {
		response.Success = common.UnmountCluster(mandatory, request.Module, request.Cluster)
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

	module := ""
	if moduleStr, found := urlMapping["module"]; found {
		module = moduleStr[0]
	}

	cluster := ""
	if clusterStr, found := urlMapping["cluster"]; found {
		cluster = clusterStr[0]
	}

	var id uint64
	if idStr, found := urlMapping["id"]; found {
		if tmp, err := strconv.ParseUint(idStr[0], 10, 64); err == nil {
			id = tmp
		} else {
			id = 0
		}
	} else {
		id = 0
	}

	if (module == "") && (cluster == "") && (id == 0) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &interfaces.HTTPResponse{Success: true}

	mandatory := common.ThreadMandatory{thread.C5, thread.ProcessorResponseTable, thread.config.Timeout}
	filter := supervisor.Filter{Module: module, Cluster: cluster, Id: id}

	instance, err := common.GetSupervisor(mandatory, filter)
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
		common.ThreadMandatory{
			thread.C5,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
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

	mandatory := common.ThreadMandatory{thread.C1, thread.DatabaseResponseTable, thread.config.Timeout}

	if r.Method == "GET" {

		clusterName, foundClusterName := urlMapping["config"]

		if foundClusterName {
			if config, found := common.GetConfigFromDatabase(mandatory, moduleName[0], clusterName[0]); found {
				bytes, _ := json.Marshal(config)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			if configs, found := common.GetConfigsFromDatabase(mandatory, moduleName[0]); found {
				bytes, _ := json.Marshal(configs)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}

	} else if r.Method == "POST" {

		err := common.StoreConfigInDatabase(mandatory, moduleName[0], *request)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
		}

	} else if r.Method == "PUT" {
		isOk := common.ReplaceConfigInDatabase(mandatory, moduleName[0], *request)
		if !isOk {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Method == "DELETE" {

		clusterName, foundClusterName := urlMapping["config"]

		if foundClusterName {
			if isOk := common.DeleteConfigInDatabase(mandatory, moduleName[0], clusterName[0]); !isOk {
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

	mandatory := common.ThreadMandatory{thread.C1, thread.DatabaseResponseTable, thread.config.Timeout}

	if r.Method == "GET" {

		moduleName, moduleNameFound := urlMapping["module"]
		clusterName, clusterNameFound := urlMapping["cluster"]

		if moduleNameFound && clusterNameFound {
			statistics, found := common.FindStatistics(mandatory, moduleName[0], clusterName[0])
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
	}

	response.Success = err == nil
	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) jobCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		thread.getJobCallback(w, r)
	} else if r.Method == http.MethodPost {
		thread.postJobCallback(w, r)
	} else if r.Method == http.MethodDelete {
		thread.deleteJobCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getJobCallback(w http.ResponseWriter, r *http.Request) {

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

	filter := &scheduler.Filter{
		Identifier: identifier,
		Module:     module,
		Cluster:    cluster,
		Interval: scheduler.Interval{
			Minute: minutes,
		}}

	var err error

	response := interfaces.HTTPResponse{}
	response.Data, err = common.GetJobs(common.ThreadMandatory{thread.C20, thread.SchedulerResponseTable, thread.config.Timeout}, filter)

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

func (thread *Thread) postJobCallback(w http.ResponseWriter, r *http.Request) {

	var job scheduler.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		fmt.Println("missing job passed to body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := interfaces.HTTPResponse{}
	if err := common.CreateJob(common.ThreadMandatory{thread.C20, thread.SchedulerResponseTable, thread.config.Timeout}, &job); err != nil {
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

func (thread *Thread) deleteJobCallback(w http.ResponseWriter, r *http.Request) {

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

	filter := &scheduler.Filter{
		Identifier: identifier,
		Module:     module,
		Cluster:    cluster,
		Interval: scheduler.Interval{
			Minute: minutes,
		}}

	var err error

	response := interfaces.HTTPResponse{}
	response.Data, err = common.GetJobs(common.ThreadMandatory{thread.C20, thread.SchedulerResponseTable, thread.config.Timeout}, filter)

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

func (thread *Thread) jobQueueCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		thread.getJobQueueCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getJobQueueCallback(w http.ResponseWriter, r *http.Request) {

	response := interfaces.HTTPResponse{}

	var err error
	response.Data, err = common.JobQueue(common.ThreadMandatory{thread.C20, thread.SchedulerResponseTable, thread.config.Timeout})

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
