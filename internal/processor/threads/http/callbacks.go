package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/processor/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/processor/threads"
	"github.com/GabeCordo/toolchain/multithreaded"
	"net/http"
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
		request.Run, &request.Pipeline, request.Metadata, thread.Config.Timeout)

	if errors.Is(err, multithreaded.NoResponseReceived) {
		w.WriteHeader(http.StatusInternalServerError)
	} else if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	response := interfaces.Response{Success: err == nil}
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

	response := interfaces.Response{Success: true}

	if request.Action == "shutdown" {
		threads.ShutdownCore(thread.Interrupt)
	}

	b, _ := json.Marshal(response)
	w.Write(b)
}

func (thread *Thread) debugStatsCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		thread.getDebugStatsCallback(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) getDebugStatsCallback(w http.ResponseWriter, r *http.Request) {

	stats, err := threads.GetProvisionerStatistics(thread.C1, thread.ProvisionerResponseTable, thread.Config.Timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		buildStatisticsPage(w, stats)
	}
}
