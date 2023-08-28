package common

import (
	"encoding/json"
	"github.com/GabeCordo/mango/core/interfaces/cluster"
	"github.com/GabeCordo/mango/core/interfaces/module"
	"net/http"
)

type HTTPRequest struct {
	Host   string            `json:"host"`
	Port   int               `json:"port"`
	Module HTTPModuleRequest `json:"module,omitempty"`
}

type HTTPModuleRequest struct {
	Name    string             `json:"name"`
	Config  module.Config      `json:"config,omitempty"`
	Mount   bool               `json:"mount,omitempty"`
	Cluster HTTPClusterRequest `json:"cluster,omitempty"`
}

type HTTPClusterRequest struct {
	Name       string            `json:"name"`
	Mount      bool              `json:"mount,omitempty"`
	Supervisor SupervisorRequest `json:"supervisor,omitempty"`
}

type HTTPSupervisorAction string

const (
	Update   HTTPSupervisorAction = "update"
	Crash                         = "crash"
	Complete                      = "complete"
)

type HTTPSupervisorRequest struct {
	Identifier uint64               `json:"identifier"`
	Action     HTTPSupervisorAction `json:"action,omitempty"`
	Statistics cluster.Statistics   `json:"statistics,omitempty"`
	Log        HTTPLogRequest       `json:"log,omitempty"`
	Cache      HTTPCacheRequest     `json:"cache,omitempty"`
}

type HTTPLogLevel string

const (
	Normal HTTPLogLevel = "normal"
	Warn                = "warn"
	Fatal               = "fatal"
)

type HTTPLogRequest struct {
	Message string       `json:"message"`
	Level   HTTPLogLevel `json:"level"`
}

type HTTPCacheRequest struct {
	Data any `json:"data"`
}

type Response struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}

func GetRequest(r *http.Request) (request *HTTPRequest, err error) {
	request = &HTTPRequest{}
	err = json.NewDecoder(r.Body).Decode(request)

	return request, err
}
