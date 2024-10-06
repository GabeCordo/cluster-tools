package api

import (
	"github.com/GabeCordo/cluster-tools/internal/core/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
)

type HTTPRequest struct {
	Host   string            `json:"host"`
	Port   int               `json:"port"`
	Module HTTPModuleRequest `json:"module,omitempty"`
}

type HTTPModuleRequest struct {
	Name    string                 `json:"name"`
	Config  processor.ModuleConfig `json:"config,omitempty"`
	Mount   bool                   `json:"mount,omitempty"`
	Cluster HTTPClusterRequest     `json:"cluster,omitempty"`
}

type HTTPClusterRequest struct {
	Name       string                `json:"name"`
	Mount      bool                  `json:"mount,omitempty"`
	Supervisor HTTPSupervisorRequest `json:"runner,omitempty"`
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
	Statistics statistic.Statistics `json:"statistics,omitempty"`
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
