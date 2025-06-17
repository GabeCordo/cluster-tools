package threads

import (
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/nonce"
)

type ProvisionerMandatory struct {
	Pipe          chan<- ProvisionerRequest
	ResponseTable *nonce.ResponseTable
	NoncePool     *nonce.Pool
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

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, provisionerThreadRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
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

	response, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	provisionerResponse := response.(ProvisionerResponse)
	return provisionerResponse.Error
}
