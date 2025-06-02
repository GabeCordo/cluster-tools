package rest

import (
	"encoding/json"
	"github.com/GabeCordo/Flock/internal/core/database/statistic"
	"github.com/GabeCordo/Flock/internal/core/processor"
	"net/http"
)

type Request struct {
	Host   string        `json:"host"`
	Port   int           `json:"port"`
	Module ModuleRequest `json:"module,omitempty"`
}

type ModuleRequest struct {
	Name    string                 `json:"name"`
	Config  processor.ModuleConfig `json:"config,omitempty"`
	Mount   bool                   `json:"mount,omitempty"`
	Cluster ClusterRequest         `json:"cluster,omitempty"`
}

type ClusterRequest struct {
	Name       string            `json:"name"`
	Mount      bool              `json:"mount,omitempty"`
	Supervisor SupervisorRequest `json:"runner,omitempty"`
}

type SupervisorAction string

const (
	Update   SupervisorAction = "update"
	Crash                     = "crash"
	Complete                  = "complete"
)

type SupervisorRequest struct {
	Identifier uint64               `json:"identifier"`
	Action     SupervisorAction     `json:"action,omitempty"`
	Statistics statistic.Statistics `json:"statistics,omitempty"`
	Log        LogRequest           `json:"log,omitempty"`
	Cache      CacheRequest         `json:"cache,omitempty"`
}

type LogLevel string

const (
	Normal LogLevel = "normal"
	Warn            = "warn"
	Fatal           = "fatal"
)

type LogRequest struct {
	Message string   `json:"message"`
	Level   LogLevel `json:"level"`
}

type CacheRequest struct {
	Data any `json:"data"`
}

type Response struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}

func GetRequest(r *http.Request) (request *Request, err error) {
	request = &Request{}
	err = json.NewDecoder(r.Body).Decode(request)

	return request, err
}
