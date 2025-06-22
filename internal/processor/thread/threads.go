package thread

import (
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/shared/nonce"
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
	ProvisionerRunCreate
	ProvisionerRunStop
	ProvisionerStatisticsGet
	ProvisionerRegisterModules
)

type ProvisionerRequest struct {
	Action     ProvisionerAction
	Namespace  string
	Supervisor uint64
	Pipeline   *pipeline.Pipeline
	Metadata   map[string]string
	Path       string
	Nonce      nonce.Nonce
}

type ProvisionerResponse struct {
	Success bool
	Error   error
	Data    any
	Nonce   nonce.Nonce
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
	Nonce  nonce.Nonce
}
