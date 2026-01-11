package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/FortifiedCode/flock/internal/targets/core/component/processor"
	"github.com/FortifiedCode/flock/internal/targets/core/database"
	"github.com/FortifiedCode/flock/internal/targets/core/database/job"
	"github.com/FortifiedCode/flock/internal/targets/core/database/pipeline"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
)

// TODO : add comments to the else conditions where the processor may support

func (t *Thread) processorCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		/* show the operator all the processors attached to the flock */
		t.getProcessorCallback(w, r)
	} else {
		/* the rest does not support any other methods on the processor */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getProcessorCallback(w http.ResponseWriter, r *http.Request) {

	processors, success := thread.GetProcessors(
		thread.Mandatory{
			Pipe:          t.channels.c5,
			Log:           t.logger,
			ResponseTable: t.ProcessorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		},
	)

	response := Response{Success: success}

	if success {
		response.Data = processors
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(response)
	_, err := w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (t *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		/* show the operator all the modules registered to the flock */
		t.getModuleCallback(w, r)
	} else if r.Method == http.MethodPut {
		/* the operator shall be allowed to mount and unmount modules */
		t.putModuleCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getModuleCallback(w http.ResponseWriter, r *http.Request) {

	success, modules := thread.GetModules(
		thread.Mandatory{
			Pipe:          t.channels.c5,
			Log:           t.logger,
			ResponseTable: t.ProcessorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		},
	)

	response := Response{Success: success}
	if success {
		response.Data = modules
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(response)
	_, err := w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
	var success = false

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c5,
		Log:           t.logger,
		ResponseTable: t.ProcessorResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	if request.Mounted {
		success, err = thread.MountModule(mandatory, request.ModuleName)
	} else {
		success, err = thread.UnmountModule(mandatory, request.ModuleName)
	}

	response := Response{Success: success}

	if errors.Is(err, processor.ModuleDoesNotExist) {
		w.WriteHeader(http.StatusNotFound)
	} else if !success {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (t *Thread) functionCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		/* the operator shall see clusters registered to the flock */
		t.getFunctionCallback(w, r)
	} else if r.Method == http.MethodPut {
		/* the operator shall mount clusters in the flock */
		/* the operator shall unmount clusters in the flock */
		t.putFunctionCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getFunctionCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	moduleName, foundModuleName := urlMapping["module"]

	if !foundModuleName {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clusterList, success := thread.GetFunctions(
		thread.Mandatory{
			Pipe:          t.channels.c5,
			Log:           t.logger,
			ResponseTable: t.ProcessorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		},
		moduleName[0],
	)
	if !success {
		w.WriteHeader(http.StatusNotFound)
	}

	response := Response{Success: success}

	if success {
		response.Data = clusterList
	}

	b, _ := json.Marshal(response)
	_, err := w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type FunctionConfigJSONBody struct {
	Module   string `json:"module"`
	Function string `json:"function"`
	Mounted  bool   `json:"mounted"`
}

func (t *Thread) putFunctionCallback(w http.ResponseWriter, r *http.Request) {

	request := &FunctionConfigJSONBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := Response{}

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c5,
		Log:           t.logger,
		ResponseTable: t.ProcessorResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	if request.Mounted {
		response.Success = thread.MountFunction(mandatory, request.Module, request.Function)
	} else {
		response.Success = thread.UnmountFunction(mandatory, request.Module, request.Function)
	}

	if !response.Success {
		w.WriteHeader(http.StatusNotFound)
	}

	b, _ := json.Marshal(response)
	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type JobConfigJSONBody struct {
	Namespace string            `json:"namespace"`
	Pipeline  string            `json:"pipeline"`
	Run       uint64            `json:"id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type RunProvisionJSONResponse struct {
	Namespace string `json:"cluster,omitempty"`
	Pipeline  string `json:"pipeline,omitempty"`
	Run       uint64 `json:"id,omitempty"`
}

func (t *Thread) runCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getRunCallback(w, r)
	} else if r.Method == http.MethodPost {
		t.postRunCallback(w, r)
	} else if r.Method == http.MethodDelete {
		t.deleteRunCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getRunCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	namespace := ""
	if namespaceStr, found := urlMapping["namespace"]; found {
		namespace = namespaceStr[0]
	}

	pipelineVar := ""
	if pipelineStr, found := urlMapping["pipeline"]; found {
		pipelineVar = pipelineStr[0]
	}

	var maximumResults uint64 = 0
	if maximumResultsStr, found := urlMapping["maximumResults"]; found {
		var err error
		maximumResults, err = strconv.ParseUint(maximumResultsStr[0], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var offsetOfResults uint64 = 0
	if offsetOfResultsStr, found := urlMapping["offsetOfResults"]; found {
		var err error
		offsetOfResults, err = strconv.ParseUint(offsetOfResultsStr[0], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var id string
	if idStr, found := urlMapping["id"]; found {
		var err error
		_, err = strconv.ParseUint(idStr[0], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			id = idStr[0]
		}
	} else {
		id = "0"
	}

	if (namespace == "") && (pipelineVar == "") && (id == "0") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &Response{Success: true}

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c5,
		Log:           t.logger,
		ResponseTable: t.ProcessorResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}
	filter := database.Filter{
		Namespace:       namespace,
		Pipeline:        pipelineVar,
		Identifier:      id,
		MaximumResults:  maximumResults,
		OffsetOfResults: offsetOfResults,
	}

	instance, err := thread.GetRun(mandatory, filter)
	if err != nil {
		response.Success = false
		response.Description = err.Error()
	} else {
		sort.Slice(instance, func(i, j int) bool {
			return instance[i].Id > instance[j].Id
		})
		response.Data = instance
	}

	b, _ := json.Marshal(response)
	_, err = w.Write(b)
	if err != nil {
		t.logger.Alert(err.Error())
	}
}

func (t *Thread) postRunCallback(w http.ResponseWriter, r *http.Request) {

	var request JobConfigJSONBody

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		t.logger.Printf("[POST][/run] didn't receive a valid body %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if runId, err := thread.CreateRun(
		thread.Mandatory{
			Pipe:          t.channels.c5,
			Log:           t.logger,
			ResponseTable: t.ProcessorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		},
		thread.HttpClient,
		request.Namespace,
		request.Pipeline,
		request.Metadata,
	); err == nil {

		response := &RunProvisionJSONResponse{
			Namespace: request.Namespace,
			Pipeline:  request.Pipeline,
			Run:       runId,
		}
		bytes, _ := json.Marshal(response)
		if _, err := w.Write(bytes); err != nil {
			// TODO : support module is not mounted
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		t.logger.Printf("[POST][/run] encountered error %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (t *Thread) deleteRunCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	runIdStr, runIdStrFound := urlMapping["id"]
	if !runIdStrFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	runId, err := strconv.ParseUint(runIdStr[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = thread.StopRun(
		thread.Mandatory{
			Pipe:          t.channels.c5,
			Log:           t.logger,
			ResponseTable: t.ProcessorResponseTable,
			NoncePool:     t.noncePool,
			Timeout:       t.config.Timeout,
		},
		runId,
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (t *Thread) runCountCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getRunCountCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getRunCountCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	namespace := ""
	if namespaceStr, found := urlMapping["namespace"]; found {
		namespace = namespaceStr[0]
	}

	pipelineVar := ""
	if pipelineStr, found := urlMapping["pipeline"]; found {
		pipelineVar = pipelineStr[0]
	}

	if (namespace == "") && (pipelineVar == "") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &Response{Success: true}

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c5,
		Log:           t.logger,
		ResponseTable: t.ProcessorResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}
	filter := database.Filter{
		Namespace: namespace,
		Pipeline:  pipelineVar,
	}

	count, err := thread.GetRunCount(mandatory, filter)
	if err != nil {
		response.Success = false
		response.Description = err.Error()
	} else {
		response.Data = count
	}

	b, _ := json.Marshal(response)
	_, err = w.Write(b)
	if err != nil {
		t.logger.Alert(err.Error())
	}
}

func (t *Thread) namespaceCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getNamespaceCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getNamespaceCallback(w http.ResponseWriter, r *http.Request) {

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c1,
		Log:           t.logger,
		ResponseTable: t.DatabaseResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	namespaces, err := thread.GetNamespacesFromDatabase(mandatory)
	response := Response{
		Success: err == nil,
		Data:    namespaces,
	}

	bytes, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (t *Thread) pipelineCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getPipelineCallback(w, r)
	} else if r.Method == http.MethodPost {
		t.postPipelineCallback(w, r)
	} else if r.Method == http.MethodPut {
		t.putPipelineCallback(w, r)
	} else if r.Method == http.MethodDelete {
		t.deletePipelineCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (t *Thread) getPipelineCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	mapping, foundMapping := urlMapping["namespace"]

	var namespaceId string
	if foundMapping {
		namespaceId = mapping[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mapping, foundMapping = urlMapping["pipeline"]

	var pipelineId string
	if foundMapping {
		pipelineId = mapping[0]
	} else {
		pipelineId = ""
	}

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c1,
		Log:           t.logger,
		ResponseTable: t.DatabaseResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	if cfg, found := thread.GetPipelineFromDatabase(mandatory, namespaceId, pipelineId); found {
		bytes, _ := json.Marshal(cfg)
		if _, err := w.Write(bytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (t *Thread) postPipelineCallback(w http.ResponseWriter, r *http.Request) {

	request := new(pipeline.Pipeline)
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if (request.Identifier == "") || (request.Namespace == "") || (request.Data == nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c1,
		Log:           t.logger,
		ResponseTable: t.DatabaseResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	err = thread.StorePipelineInDatabase(mandatory, request.Namespace, request.Identifier, request.Data)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
	}
}

func (t *Thread) putPipelineCallback(w http.ResponseWriter, r *http.Request) {

	request := new(pipeline.Pipeline)
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c1,
		Log:           t.logger,
		ResponseTable: t.DatabaseResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	isOk := thread.ReplacePipelineInDatabase(mandatory, request.Namespace, request.Identifier, request.Data)
	if !isOk {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (t *Thread) deletePipelineCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	mapping, foundMapping := urlMapping["namespace"]

	var namespaceId string
	if foundMapping {
		namespaceId = mapping[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mapping, foundMapping = urlMapping["pipeline"]

	var pipelineId string
	if foundMapping {
		pipelineId = mapping[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c1,
		Log:           t.logger,
		ResponseTable: t.DatabaseResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	if isOk := thread.DeletePipelineInDatabase(mandatory, namespaceId, pipelineId); !isOk {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (t *Thread) statisticInfoCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getStatisticInfoCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getStatisticInfoCallback(w http.ResponseWriter, r *http.Request) {

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c1,
		Log:           t.logger,
		ResponseTable: t.DatabaseResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	var namespace string
	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	namespaceName, foundNamespaceName := urlMapping["namespace"]
	if foundNamespaceName {
		namespace = namespaceName[0]
	} else {
		namespace = ""
	}

	fields, err := thread.FindStatistics(mandatory, namespace)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(fields)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(bytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (t *Thread) statisticCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getStatisticCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getStatisticCallback(w http.ResponseWriter, r *http.Request) {

	mandatory := thread.Mandatory{
		Pipe:          t.channels.c1,
		Log:           t.logger,
		ResponseTable: t.DatabaseResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	var namespaceName string
	namespaceSlice, namespaceNameFound := urlMapping["namespace"]
	if namespaceNameFound {
		namespaceName = namespaceSlice[0]
	} else {
		namespaceName = ""
	}

	var pipelineName string
	pipelineSlice, pipelineNameFound := urlMapping["pipeline"]
	if pipelineNameFound {
		pipelineName = pipelineSlice[0]
	} else {
		pipelineName = ""
	}

	statistics, found := thread.FindStatistic(mandatory, namespaceName, pipelineName)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(statistics)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(bytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (t *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		t.getDebugCallback(w, r)
	} else if r.Method == http.MethodPost {
		t.postDebugCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getDebugCallback(w http.ResponseWriter, r *http.Request) {

	t.logger.Printf("ping from %s\n", r.RemoteAddr)
	response := Response{Success: true, Description: "bonjour"}
	b, _ := json.Marshal(response)
	_, err := w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
		t.logger.Println("missing body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := Response{Success: true}

	if request.Action == "shutdown" {
		err = thread.ShutdownCore(t.channels.interrupt)
		if err != nil {
			response.Description = err.Error()
		}
	}

	response.Success = err == nil
	b, _ := json.Marshal(response)
	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
		Namespace:  module,
		Pipeline:   cluster,
		Interval: database.Interval{
			Minute: minutes,
		}}

	var err error

	response := Response{}
	response.Data, err = thread.GetJobs(thread.Mandatory{
		Pipe:          t.channels.c20,
		Log:           t.logger,
		ResponseTable: t.SchedulerResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	}, filter)

	response.Success = err == nil
	if err != nil {
		response.Description = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (t *Thread) postJobCallback(w http.ResponseWriter, r *http.Request) {

	var j job.Job
	err := json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		t.logger.Println("missing job passed to body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := Response{}
	err = thread.CreateJob(thread.Mandatory{
		Pipe:          t.channels.c20,
		Log:           t.logger,
		ResponseTable: t.SchedulerResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout},
		&j)
	if err != nil {
		response.Success = false
		response.Data = err.Error()
	} else {
		response.Success = true
	}

	var b []byte
	b, err = json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		Namespace:  module,
		Pipeline:   cluster,
		Interval: database.Interval{
			Minute: minutes,
		}}

	var err error

	response := Response{}
	err = thread.DeleteJob(thread.Mandatory{
		Pipe:          t.channels.c20,
		Log:           t.logger,
		ResponseTable: t.SchedulerResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout}, filter)

	response.Success = err == nil
	if err != nil {
		response.Description = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

	response := Response{}

	var err error
	response.Data, err = thread.JobQueue(thread.Mandatory{
		Pipe:          t.channels.c20,
		Log:           t.logger,
		ResponseTable: t.SchedulerResponseTable,
		NoncePool:     t.noncePool,
		Timeout:       t.config.Timeout,
	})

	if err != nil {
		response.Description = err.Error()
	}
	response.Success = err == nil

	b, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
