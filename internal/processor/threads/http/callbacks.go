package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/processor/api"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
	"net/url"
	"time"
)

type JSONResponse struct {
	Status      int    `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Data        any    `json:"data,omitempty"`
}

type RunConfigJSONBody struct {
	Namespace string            `json:"module"`
	Pipeline  pipeline.Pipeline `json:"pipeline"`
	Run       uint64            `json:"id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type RunProvisionJSONResponse struct {
	Namespace string `json:"cluster,omitempty"`
	Pipeline  string `json:"pipeline,omitempty"`
	Run       uint64 `json:"id,omitempty"`
}

func (thread *Thread) runCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		thread.postRunCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) postRunCallback(w http.ResponseWriter, r *http.Request) {

	var request RunConfigJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = threads.RunProvision(thread.C1, thread.ProvisionerResponseTable, request.Namespace,
		request.Run, &request.Pipeline, request.Metadata, *thread.Config.Timeout)

	if errors.Is(err, multithreaded.NoResponseReceived) {
		w.WriteHeader(http.StatusInternalServerError)
	} else if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	response := api.Response{Success: err == nil}
	if err != nil {
		response.Description = err.Error()
	}
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
	// do nothing
}

func (thread *Thread) postDebugCallback(w http.ResponseWriter, r *http.Request) {

	request := &DebugJSONBody{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if (r.Method != "OPTIONS") && err != nil {
		fmt.Println("missing body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := api.Response{Success: true}

	if request.Action == "shutdown" {
		threads.ShutdownCore(thread.Interrupt)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) gateCallback(w http.ResponseWriter, r *http.Request) {

	// TODO: use DELETE and PUT HTTP directives
	if r.Method == http.MethodGet {
		thread.getGateCallback(w, r)
	} else if r.Method == http.MethodPost {
		thread.postGateCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getGateCallback(w http.ResponseWriter, r *http.Request) {

	thread.mutex.RLock()
	defer thread.mutex.RUnlock()

	if *thread.Config.Standalone {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := api.Response{Success: true, Data: *thread.Config.Core}
	json.NewEncoder(w).Encode(response)
}

func (thread *Thread) postGateCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	coreHost, coreHostFound := urlMapping["core"]
	if !coreHostFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	action, actionFound := urlMapping["action"]
	if !actionFound || !((action[0] == "connect") || (action[0] == "disconnect")) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	thread.mutex.Lock()
	defer thread.mutex.Unlock()

	if action[0] == "connect" {

		// the processor shall disconnect from an ongoing gateway connection
		// before attempting to connect to another
		//
		// relation: [core] 1-* [processor]
		//
		if !*thread.Config.Standalone {
			err := api.DisconnectFromCore(*thread.Config.Core, &thread.Config.ExternalNet)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}
		}

		*thread.Config.Core = coreHost[0]
		*thread.Config.Standalone = false

		err := api.ConnectToCore(*thread.Config.Core, &thread.Config.ExternalNet)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {

		if *thread.Config.Standalone {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := api.DisconnectFromCore(*thread.Config.Core, &thread.Config.ExternalNet)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func (thread *Thread) debugStatsCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		thread.getDebugStatsCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getDebugStatsCallback(w http.ResponseWriter, r *http.Request) {

	stats, err := threads.GetProvisionerStatistics(thread.C1, thread.ProvisionerResponseTable, *thread.Config.Timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		if b, err := json.Marshal(stats); err == nil {
			w.Write(b)
		}
	}
}
