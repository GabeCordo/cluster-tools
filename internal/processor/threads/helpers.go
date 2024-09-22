package threads

import (
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/interfaces"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

func SupervisorProvision(pipe chan<- ProvisionerRequest, responseTable *multithreaded.ResponseTable,
	moduleName, clusterName string, supervisor uint64, meta map[string]string, cfg *cluster.Config, timeout float64) error {

	// there is a possibility the user never passed an args value to the HTTP endpoint,
	// so we need to replace it with and empty array
	if meta == nil {
		meta = make(map[string]string)
	}
	provisionerThreadRequest := ProvisionerRequest{
		Action:     ProvisionerSupervisorCreate,
		Module:     moduleName,
		Cluster:    clusterName,
		Supervisor: supervisor,
		Metadata:   meta,
		Config:     cfg,
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

func ShutdownCore(pipe chan<- InterruptEvent) {
	pipe <- Shutdown
}

func GetProvisionerStatistics(pipe chan<- ProvisionerRequest, responseTable *multithreaded.ResponseTable,
	timeout float64) ([]*interfaces.SupervisorSummary, error) {

	request := ProvisionerRequest{
		Action: ProvisionerStatisticsGet,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	rsp := (data).(ProvisionerResponse)

	collectedStatistics := (rsp.Data).([]*interfaces.SupervisorSummary)
	return collectedStatistics, nil
}
