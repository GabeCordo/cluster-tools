package thread

import (
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/shared/nonce"
)

type ProvisionerMandatory struct {
	Pipe      chan<- *ProvisionerRequest
	NoncePool *nonce.Pool
}

func AsyncRunStart(mandatory ProvisionerMandatory, namespace string, supervisor uint64, cfg *pipeline.Pipeline, meta map[string]string) {

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
