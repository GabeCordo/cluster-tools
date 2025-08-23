package thread

import (
	"github.com/FortifiedCode/flock/internal/shared/nonce"
	"github.com/FortifiedCode/plover"
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
	Pipeline   *plover.PipelineIR
	Metadata   map[string]string
	Path       string
	Nonce      nonce.Nonce
}

func NewProvisionerRequest() *ProvisionerRequest {
	r := new(ProvisionerRequest)
	if r == nil {
		panic("failed to allocate ProvisionerRequest")
	}
	return r
}

type ProvisionerResponse struct {
	Success bool
	Error   error
	Data    any
	Nonce   nonce.Nonce
}

func NewProvisionerResponse(req *ProvisionerRequest) *ProvisionerResponse {
	rsp := new(ProvisionerResponse)
	if rsp == nil {
		panic("failed to allocate ProvisionerResponse")
	}
	rsp.Nonce = req.Nonce
	return rsp
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

func NewSocketRequest() *SocketRequest {
	r := new(SocketRequest)
	if r == nil {
		panic("failed to allocate SocketRequest")
	}
	return r
}
