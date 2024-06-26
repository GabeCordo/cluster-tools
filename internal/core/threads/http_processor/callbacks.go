package http_processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/components/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
	"net/url"
	"strconv"
)

func (thread *Thread) processorCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		/* the operator wants to register a new processor to the core */
		thread.postProcessorCallback(w, r)
	} else if r.Method == "DELETE" {
		/* the operator wants to delete a processor from the server */
		thread.deleteProcessorCallback(w, r)
	} else {
		/* we don't support the method for this resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) postProcessorCallback(w http.ResponseWriter, r *http.Request) {

	request, err := interfaces.GetRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	cfg := &interfaces.ProcessorConfig{Host: request.Host, Port: request.Port}
	success, err := common.AddProcessor(
		common.ThreadMandatory{
			thread.C7,
			thread.ProcessorResponseTable,
			thread.config.Timeout},
		cfg,
	)

	if errors.Is(err, processor.AlreadyExists) {
		w.WriteHeader(http.StatusConflict)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	response := interfaces.HTTPResponse{Success: success}
	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) deleteProcessorCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	hostName, hostNameFound := urlMapping["host"]
	if !hostNameFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	portStr, portFound := urlMapping["port"]
	if !portFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	port, err := strconv.Atoi(portStr[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg := &interfaces.ProcessorConfig{Host: hostName[0], Port: port}
	err = common.DeleteProcessor(
		common.ThreadMandatory{
			thread.C7,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
		cfg,
	)

	response := interfaces.HTTPResponse{Success: err == nil}

	if errors.Is(err, processor.DoesNotExist) {
		w.WriteHeader(http.StatusNotFound)
	} else if errors.Is(err, multithreaded.NoResponseReceived) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		/* operator wishes to add a new module to a processor */
		thread.postModuleCallback(w, r)
	} else if r.Method == "DELETE" {
		/* operator wishes to remove an existing module from a processor */
		thread.deleteModuleCallback(w, r)
	} else {
		/* the method is not supported for this resource type */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) postModuleCallback(w http.ResponseWriter, r *http.Request) {

	request, err := interfaces.GetRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	processorName := fmt.Sprintf("%s:%d", request.Host, request.Port)
	success, err := common.AddModule(
		common.ThreadMandatory{
			thread.C7,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
		processorName,
		&request.Module.Config,
	)

	response := interfaces.HTTPResponse{Success: success}

	if !success && errors.Is(err, processor.ModuleAlreadyRegistered) {
		/* the module is already registered to the processor */
		w.WriteHeader(http.StatusConflict)
	} else if !success && errors.Is(err, processor.DoesNotExist) {
		/* the processor does not exist and can not bind a module */
		w.WriteHeader(http.StatusBadRequest)
	} else if !success {
		w.WriteHeader(http.StatusBadRequest)
	}

	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) deleteModuleCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	hostName, hostNameFound := urlMapping["host"]
	if !hostNameFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	portStr, portFound := urlMapping["port"]
	if !portFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	port, err := strconv.Atoi(portStr[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	moduleName, moduleNameFound := urlMapping["module"]
	if !moduleNameFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = common.DeleteModule(
		common.ThreadMandatory{
			thread.C7,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
		hostName[0],
		port,
		moduleName[0],
	)

	response := interfaces.HTTPResponse{Success: err == nil}

	if errors.Is(err, processor.DoesNotExist) || errors.Is(err, processor.ModuleDoesNotExist) {
		w.WriteHeader(http.StatusNotFound)
	} else if errors.Is(err, multithreaded.NoResponseReceived) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) cacheCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		/* the program wants to grab an existing cached value */
		thread.getCacheCallback(w, r)
	} else if r.Method == "POST" {
		/* the program wants to create a new cached value */
		thread.postCacheCallback(w, r)
	} else if r.Method == "PUT" {
		/* the program wants to swap an existing cached value */
		thread.putCacheCallback(w, r)
	} else {
		/* the endpoint does not support any other methods on the resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getCacheCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	key, keyFound := urlMapping["key"]

	if !keyFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, found := common.FetchFromCache(
		common.ThreadMandatory{
			thread.C9,
			thread.CacheResponseTable,
			thread.config.Timeout,
		},
		key[0],
	)

	response := interfaces.HTTPResponse{Success: found}

	if found {
		response.Data = value
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

type CacheBody struct {
	Value  any     `json:"value"`
	Expiry float64 `json:"expiry"`
	Key    string  `json:"key,omitempty"`
}

func (thread *Thread) postCacheCallback(w http.ResponseWriter, r *http.Request) {

	request := &CacheBody{}
	json.NewDecoder(r.Body).Decode(request)

	expiry := thread.config.Timeout
	if request.Expiry != 0.0 {
		expiry = request.Expiry
	}

	identifier, success := common.StoreInCache(
		common.ThreadMandatory{
			thread.C9,
			thread.CacheResponseTable,
			thread.config.Timeout,
		},
		request.Value,
		expiry,
	)

	response := interfaces.HTTPResponse{Success: success, Data: identifier}
	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) putCacheCallback(w http.ResponseWriter, r *http.Request) {

	request := &CacheBody{}
	json.NewDecoder(r.Body).Decode(request)

	success := common.SwapInCache(
		common.ThreadMandatory{
			thread.C9,
			thread.CacheResponseTable,
			thread.config.Timeout,
		},
		request.Key,
		request.Value,
	)

	response := interfaces.HTTPResponse{Success: success}

	if !success {
		w.WriteHeader(http.StatusNotFound)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) logCallback(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method == "POST" {
		/* the program wants to log a new event */
		thread.postLogCallback(w, r)
	} else {
		/* the endpoint does not support any other methods on the resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) postLogCallback(w http.ResponseWriter, r *http.Request) {

	log := &supervisor.Log{}
	err := json.NewDecoder(r.Body).Decode(log)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = common.Log(
		common.ThreadMandatory{
			thread.C7,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
		log,
	)

	response := interfaces.HTTPResponse{Success: err == nil}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) supervisorCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		/* the processor requests to update a provisioned supervisor */
		thread.putSupervisorCallback(w, r)
	} else {
		/* the http_processor cannot call any other methods on this resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) putSupervisorCallback(w http.ResponseWriter, r *http.Request) {

	instance := &supervisor.Supervisor{}
	err := json.NewDecoder(r.Body).Decode(instance)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &interfaces.HTTPResponse{}
	err = common.UpdateSupervisor(
		common.ThreadMandatory{
			thread.C7,
			thread.ProcessorResponseTable,
			thread.config.Timeout,
		},
		instance,
	)

	response.Success = err == nil
	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		// treat this as a probe to the server
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
