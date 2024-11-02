package threads

import (
	"github.com/Sentmint/PipelineOps/internal/core/database/pipeline"
)

type InterruptEvent uint8

const (
	Shutdown InterruptEvent = 0
	Panic                   = 1
)

type ProvisionerConfig struct {
	Debug string
}

type ProvisionerAction uint8

const (
	ProvisionerModuleGet ProvisionerAction = iota
	ProvisionerRunGet
	ProvisionerRunCreate
	ProvisionerRunStop
	ProvisionerStatisticsGet
	ProvisionerRegisterModules
)

type ProvisionerSource string

const (
	Core ProvisionerSource = "core"
	User                   = "user"
)

type ProvisionerRequest struct {
	Action     ProvisionerAction
	Source     ProvisionerSource
	Namespace  string
	Supervisor uint64
	Pipeline   *pipeline.Pipeline
	Metadata   map[string]string
	Path       string
	Nonce      uint32
}

type ProvisionerResponse struct {
	Success bool
	Error   error
	Data    any
	Nonce   uint32
}

type SocketAction uint8

const (
	SocketModuleAdd SocketAction = iota
	SocketLogAdd
	SocketRunUpdate
)

type SocketRequest struct {
	Action SocketAction
	Data   any
	Nonce  uint32
}
