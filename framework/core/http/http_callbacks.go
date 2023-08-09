package http

import (
	"encoding/json"
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl/framework/core/common"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ModuleRequestBody struct {
	ModulePath string `json:"path,omitempty"`
	ModuleName string `json:"module,omitempty"`
}

func (httpThread *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

	request := &ModuleRequestBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if (r.Method != "GET") && (err != nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		success, modules := common.GetModules(httpThread.C5, httpThread.ProvisionerResponseTable)
		if !success {
			w.WriteHeader(http.StatusInternalServerError)
		}
		moduleBytes, err := json.Marshal(modules)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = w.Write(moduleBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		if success, description := common.RegisterModule(httpThread.C5, httpThread.ProvisionerResponseTable, request.ModulePath); !success {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(description))
		}
	} else if r.Method == "DELETE" {
		success, description := common.DeleteModule(httpThread.C5, httpThread.ProvisionerResponseTable, request.ModuleName)
		if !success {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte(description))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type JSONResponse struct {
	Status      int    `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Data        any    `json:"data,omitempty"`
}

type ClusterConfigJSONBody struct {
	Module  string `json:"module"`
	Cluster string `json:"cluster"`
	Mounted bool   `json:"mounted"`
}

func (httpThread *Thread) clusterCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	request := &ClusterConfigJSONBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if (r.Method != "GET") && (err != nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		if moduleName, foundModuleName := urlMapping["module"]; foundModuleName {
			clusterList, success := common.GetClusters(httpThread.C5, httpThread.ProvisionerResponseTable, moduleName[0])
			if !success {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			clusterBytes, err := json.Marshal(clusterList)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = w.Write(clusterBytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else if r.Method == "PUT" {
		if request.Mounted {
			success := common.ClusterMount(httpThread.C5, httpThread.ProvisionerResponseTable, request.Module, request.Cluster)
			if !success {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			success := common.ClusterUnMount(httpThread.C5, httpThread.ProvisionerResponseTable, request.Module, request.Cluster)
			if !success {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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

func (httpThread *Thread) supervisorCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	var request SupervisorConfigJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if (r.Method != "GET") && (err != nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {

		clusterName, foundClusterName := urlMapping["cluster"]
		moduleName, foundModuleName := urlMapping["module"]
		supervisorIdStr, foundSupervisorId := urlMapping["id"]

		if !foundClusterName || !foundModuleName {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			if foundSupervisorId {
				supervisorId, err := strconv.ParseUint(supervisorIdStr[0], 10, 64)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				supervisorInst, success := common.GetSupervisor(httpThread.C5, httpThread.ProvisionerResponseTable,
					moduleName[0], clusterName[0], supervisorId)
				if !success {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				supervisorBytes, err := json.Marshal(supervisorInst)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				_, err = w.Write(supervisorBytes)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				supervisors, success := common.GetSupervisors(httpThread.C5, httpThread.ProvisionerResponseTable, moduleName[0], clusterName[0])
				if !success {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				supervisorsBytes, err := json.Marshal(supervisors)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				_, err = w.Write(supervisorsBytes)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}
	} else if r.Method == "POST" {
		if supervisorId, success, description := common.SupervisorProvision(
			httpThread.C5,
			httpThread.ProvisionerResponseTable,
			request.Module,
			request.Cluster,
			request.Metadata,
			request.Config); success {

			response := &SupervisorProvisionJSONResponse{Cluster: request.Cluster, Supervisor: supervisorId}
			bytes, _ := json.Marshal(response)
			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(description))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (httpThread *Thread) configCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	request := &cluster.Config{}
	err := json.NewDecoder(r.Body).Decode(request)
	if (r.Method != "GET") && (err != nil) {
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

		clusterName, foundClusterName := urlMapping["cluster"]

		if foundClusterName {
			if config, found := common.GetConfigFromDatabase(httpThread.C1, httpThread.DatabaseResponseTable, moduleName[0], clusterName[0]); found {
				bytes, _ := json.Marshal(config)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	} else if r.Method == "POST" {

		isOk := common.StoreConfigInDatabase(httpThread.C1, httpThread.DatabaseResponseTable, moduleName[0], *request)
		if !isOk {
			w.WriteHeader(http.StatusConflict)
		}

	} else if r.Method == "PUT" {
		isOk := common.ReplaceConfigInDatabase(httpThread.C1, httpThread.DatabaseResponseTable, moduleName[0], *request)
		if !isOk {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (httpThread *Thread) statisticCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	if r.Method == "GET" {

		moduleName, moduleNameFound := urlMapping["module"]
		clusterName, clusterNameFound := urlMapping["cluster"]

		if moduleNameFound && clusterNameFound {
			statistics, found := common.FindStatistics(httpThread.C1, httpThread.DatabaseResponseTable, moduleName[0], clusterName[0])
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

func (httpThread *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	var request DebugJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		if request.Action == "shutdown" {
			common.ShutdownCore(httpThread.Interrupt)
		} else if request.Action == "ping" {
			startTime := time.Now()
			success := common.PingNodeChannels(httpThread.logger, httpThread.C1, httpThread.DatabaseResponseTable, httpThread.C5, httpThread.ProvisionerResponseTable)
			response := DebugJSONResponse{Success: success, Duration: time.Now().Sub(startTime)}
			bytes, err := json.Marshal(response)
			if err == nil {
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else if request.Action == "debug" {
			description := common.ToggleDebugMode(httpThread.logger)
			if _, err := w.Write([]byte(description)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
