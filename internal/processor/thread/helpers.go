package thread

import (
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	nonce2 "github.com/GabeCordo/Flock/internal/shared/nonce"
)

type ProvisionerMandatory struct {
	Pipe          chan<- ProvisionerRequest
	ResponseTable *nonce2.ResponseTable
	NoncePool     *nonce2.Pool
	Timeout       float64
}

func RunStart(mandatory ProvisionerMandatory, namespace string, supervisor uint64, cfg *pipeline.Pipeline, meta map[string]string) error {

	// there is a possibility the user never passed an args value to the HTTP endpoint,
	// so we need to replace it with and empty array
	if meta == nil {
		meta = make(map[string]string)
	}
	provisionerThreadRequest := ProvisionerRequest{
		Action:     ProvisionerRunCreate,
		Namespace:  namespace,
		Supervisor: supervisor,
		Pipeline:   cfg,
		Metadata:   meta,
		Nonce:      mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- provisionerThreadRequest

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, provisionerThreadRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
	}

	provisionerResponse := (data).(ProvisionerResponse)
	return provisionerResponse.Error
}

func RunStop(mandatory ProvisionerMandatory, run uint64) error {

	request := ProvisionerRequest{
		Action:     ProvisionerRunStop,
		Supervisor: run,
		Nonce:      mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	response, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
	}

	provisionerResponse := response.(ProvisionerResponse)
	return provisionerResponse.Error
}
