package thread

import (
	"github.com/GabeCordo/DistributedFunctions/internal/shared/nonce"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/run"
	"github.com/GabeCordo/ScalingFunctions"
)

type ProvisionerMandatory struct {
	Pipe      chan<- *ProvisionerRequest
	NoncePool *nonce.Pool
}

func AsyncRunStart(mandatory ProvisionerMandatory, namespace string, supervisor uint64, cfg *ScalingFunctions.PipelineIR, meta map[string]string) {

	// there is a possibility the user never passed an args value to the HTTP endpoint,
	// so we need to replace it with and empty array
	if meta == nil {
		meta = make(map[string]string)
	}
	provisionerThreadRequest := NewProvisionerRequest()

	provisionerThreadRequest.Action = ProvisionerRunCreate
	provisionerThreadRequest.Namespace = namespace
	provisionerThreadRequest.Supervisor = supervisor
	provisionerThreadRequest.Pipeline = cfg
	provisionerThreadRequest.Metadata = meta
	provisionerThreadRequest.Nonce = mandatory.NoncePool.Next()

	mandatory.Pipe <- provisionerThreadRequest
}

func AsyncRunStop(mandatory ProvisionerMandatory, run uint64) {

	request := NewProvisionerRequest()

	request.Action = ProvisionerRunStop
	request.Supervisor = run
	request.Nonce = mandatory.NoncePool.Next()

	mandatory.Pipe <- request
}

func AsyncRequestModules(mandatory ProvisionerMandatory) {

	request := NewProvisionerRequest()
	request.Action = ProvisionerRegisterModules
	request.Nonce = mandatory.NoncePool.Next()

	mandatory.Pipe <- request
}

type SocketMandatory struct {
	Pipe      chan<- *SocketRequest
	NoncePool *nonce.Pool
}

func AsyncRunUpdate(mandatory SocketMandatory, run *run.Run) {

	sR := NewSocketRequest()
	sR.Action = SocketRunUpdate
	sR.Data = run
	sR.Nonce = mandatory.NoncePool.Next()
	mandatory.Pipe <- sR
}

func AsyncModuleAdd(mandatory SocketMandatory, moduleIR *ScalingFunctions.ModuleIR) {

	sR := NewSocketRequest()
	sR.Action = SocketModuleAdd
	sR.Data = *moduleIR
	sR.Nonce = mandatory.NoncePool.Next()
	mandatory.Pipe <- sR
}
