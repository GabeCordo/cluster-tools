package threads

import (
	"github.com/GabeCordo/toolchain/multithreaded"
	"github.com/Sentmint/yule"
	"math/rand"
)

func RunProvision(pipe chan<- ProvisionerRequest, responseTable *multithreaded.ResponseTable,
	namespace string, supervisor uint64, cfg *yule.Pipeline, meta map[string]string, timeout float64) error {

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
		Nonce:      rand.Uint32(),
	}
	pipe <- provisionerThreadRequest

	data, didTimeout := multithreaded.SendAndWait(responseTable, provisionerThreadRequest.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	provisionerResponse := (data).(ProvisionerResponse)
	return provisionerResponse.Error
}

func RunStop(pipe chan<- ProvisionerRequest, responseTable *multithreaded.ResponseTable,
	run uint64, timeout float64) error {

	request := ProvisionerRequest{
		Action:     ProvisionerRunStop,
		Supervisor: run,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	response, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	provisionerResponse := response.(ProvisionerResponse)
	return provisionerResponse.Error
}

func ShutdownCore(pipe chan<- InterruptEvent) {
	pipe <- Shutdown
}
