package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/core/message/log"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
	"github.com/GabeCordo/cluster-tools/internal/core/thread/rest"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
	"net/url"
	"strconv"
)

func (t *Thread) processorCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		/* the operator wants to register a new processor to the cluster-tools */
		t.postProcessorCallback(w, r)
	} else if r.Method == "DELETE" {
		/* the operator wants to delete a processor from the server */
		t.deleteProcessorCallback(w, r)
	} else {
		/* we don't support the method for this resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) postProcessorCallback(w http.ResponseWriter, r *http.Request) {

	request, err := rest.GetRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg := &processor.Config{Host: request.Host, Port: request.Port}
	success, err := thread.AddProcessor(
		thread.Mandatory{
			t.C7,
			t.ProcessorResponseTable,
			t.config.Timeout},
		cfg,
	)

	if errors.Is(err, processor.AlreadyExists) {
		w.WriteHeader(http.StatusConflict)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	response := rest.Response{Success: success}
	if err != nil {
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (t *Thread) deleteProcessorCallback(w http.ResponseWriter, r *http.Request) {

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

	cfg := &processor.Config{Host: hostName[0], Port: port}
	err = thread.DeleteProcessor(
		thread.Mandatory{
			t.C7,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
		cfg,
	)

	response := rest.Response{Success: err == nil}

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

func (t *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		/* operator wishes to add a new module to a processor */
		t.postModuleCallback(w, r)
	} else if r.Method == "DELETE" {
		/* operator wishes to remove an existing module from a processor */
		t.deleteModuleCallback(w, r)
	} else {
		/* the method is not supported for this resource type */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) postModuleCallback(w http.ResponseWriter, r *http.Request) {

	request, err := rest.GetRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	processorName := fmt.Sprintf("%s:%d", request.Host, request.Port)
	success, err := thread.AddModule(
		thread.Mandatory{
			t.C7,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
		processorName,
		&request.Module.Config,
	)

	response := rest.Response{Success: success}

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

func (t *Thread) deleteModuleCallback(w http.ResponseWriter, r *http.Request) {

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

	_, err = thread.DeleteModule(
		thread.Mandatory{
			t.C7,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
		hostName[0],
		port,
		moduleName[0],
	)

	response := rest.Response{Success: err == nil}

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

func (t *Thread) cacheCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		/* the program wants to grab an existing cached value */
		t.getCacheCallback(w, r)
	} else if r.Method == "POST" {
		/* the program wants to create a new cached value */
		t.postCacheCallback(w, r)
	} else if r.Method == "PUT" {
		/* the program wants to swap an existing cached value */
		t.putCacheCallback(w, r)
	} else {
		/* the endpoint does not support any other methods on the resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) getCacheCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)
	key, keyFound := urlMapping["key"]

	if !keyFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, found := thread.FetchFromCache(
		thread.Mandatory{
			t.C9,
			t.CacheResponseTable,
			t.config.Timeout,
		},
		key[0],
	)

	response := rest.Response{Success: found}

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

func (t *Thread) postCacheCallback(w http.ResponseWriter, r *http.Request) {

	request := &CacheBody{}
	json.NewDecoder(r.Body).Decode(request)

	expiry := t.config.Timeout
	if request.Expiry != 0.0 {
		expiry = request.Expiry
	}

	identifier, success := thread.StoreInCache(
		thread.Mandatory{
			t.C9,
			t.CacheResponseTable,
			t.config.Timeout,
		},
		request.Value,
		expiry,
	)

	response := rest.Response{Success: success, Data: identifier}
	b, _ := json.Marshal(response)
	w.Write(b)
}

func (t *Thread) putCacheCallback(w http.ResponseWriter, r *http.Request) {

	request := &CacheBody{}
	json.NewDecoder(r.Body).Decode(request)

	success := thread.SwapInCache(
		thread.Mandatory{
			t.C9,
			t.CacheResponseTable,
			t.config.Timeout,
		},
		request.Key,
		request.Value,
	)

	response := rest.Response{Success: success}

	if !success {
		w.WriteHeader(http.StatusNotFound)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (t *Thread) logCallback(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method == "POST" {
		/* the program wants to log a new event */
		t.postLogCallback(w, r)
	} else {
		/* the endpoint does not support any other methods on the resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) postLogCallback(w http.ResponseWriter, r *http.Request) {

	l := &log.Log{}
	err := json.NewDecoder(r.Body).Decode(l)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = thread.Log(
		thread.Mandatory{
			t.C7,
			t.ProcessorResponseTable,
			t.config.Timeout,
		},
		l,
	)

	response := rest.Response{Success: err == nil}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Description = err.Error()
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (t *Thread) supervisorCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		/* the processor requests to update a provisioned supervisor */
		t.putSupervisorCallback(w, r)
	} else {
		/* the processor cannot call any other methods on this resource */
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t *Thread) putSupervisorCallback(w http.ResponseWriter, r *http.Request) {

	instance := &supervisor.Supervisor{}
	err := json.NewDecoder(r.Body).Decode(instance)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &rest.Response{}
	err = thread.UpdateSupervisor(
		thread.Mandatory{
			t.C7,
			t.ProcessorResponseTable,
			t.config.Timeout,
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

func (t *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		// treat this as a probe to the server
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
