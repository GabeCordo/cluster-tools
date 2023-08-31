package common

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/mango/core/components/database"
	"github.com/GabeCordo/mango/core/components/processor"
	"github.com/GabeCordo/mango/core/components/supervisor"
	"github.com/GabeCordo/mango/core/interfaces/cluster"
	"github.com/GabeCordo/mango/core/interfaces/module"
	processor_i "github.com/GabeCordo/mango/core/interfaces/processor"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

func GetConfigFromDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName, clusterName string, maxWaitForResponse float64) (conf cluster.Config, found bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseFetch,
		Type:    ClusterConfig,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		maxWaitForResponse)
	if didTimeout {
		return cluster.Config{}, false
	}

	databaseResponse := (data).(DatabaseResponse)

	if !databaseResponse.Success {
		return cluster.Config{}, false
	}

	configs := (databaseResponse.Data).([]cluster.Config)
	if len(configs) != 1 {
		return cluster.Config{}, false
	}

	return configs[0], true
}

func GetConfigsFromDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName string, maxWaitForResponse float64) (configs []cluster.Config, found bool) {

	databaseRequest := DatabaseRequest{
		Action: DatabaseFetch,
		Type:   ClusterConfig,
		Module: moduleName,
		Nonce:  rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		maxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(DatabaseResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	configs = (databaseResponse.Data).([]cluster.Config)
	return configs, true
}

func StoreConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName string, cfg cluster.Config, maxWaitForResponse float64) error {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseStore,
		Type:    ClusterConfig,
		Module:  moduleName,
		Cluster: cfg.Identifier,
		Data:    cfg,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		maxWaitForResponse)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	databaseResponse := (data).(DatabaseResponse)
	// TODO : make the database generate the errors
	if !databaseResponse.Success {
		return errors.New("could not store config in database")
	}

	return nil
}

func ReplaceConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName string, cfg cluster.Config, maxWaitForResponse float64) (success bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseReplace,
		Type:    ClusterConfig,
		Module:  moduleName,
		Cluster: cfg.Identifier,
		Data:    cfg,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		maxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(DatabaseResponse)
	return databaseResponse.Success
}

func DeleteConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName, configName string, maxWaitForResponse float64) (success bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseDelete,
		Type:    ClusterConfig,
		Module:  moduleName,
		Cluster: configName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(databaseResponseTable, databaseRequest.Nonce, maxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(DatabaseResponse)
	return databaseResponse.Success
}

func GetProcessors(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	maxWaitForResponse float64) ([]processor.Processor, bool) {

	request := ProcessorRequest{
		Action: ProcessorGet,
		Source: HttpClient,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	response := (data).(ProcessorResponse)

	if response.Success {
		return (response.Data).([]processor.Processor), true
	} else {
		return nil, false
	}
}

func AddProcessor(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	cfg *processor_i.Config, maxWaitForResponse float64) (bool, error) {

	request := ProcessorRequest{
		Action: ProcessorAdd,
		Source: HttpProcessor,
		Data:   *cfg,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)
	return response.Success, response.Error
}

func DeleteProcessor(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	cfg *processor_i.Config, maxWaitForResponse float64) error {

	request := ProcessorRequest{
		Action: ProcessorRemove,
		Source: HttpProcessor,
		Data:   *cfg,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (data).(ProcessorResponse)
	return response.Error
}

func MountCluster(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	moduleName, clusterName string, maxWaitForResponse float64) (success bool) {

	request := ProcessorRequest{
		Action:      ProcessorClusterMount,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(ProcessorResponse)
	return provisionerResponse.Success
}

func UnmountCluster(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	moduleName, clusterName string, maxWaitForResponse float64) (success bool) {

	request := ProcessorRequest{
		Action:      ProcessorClusterUnmount,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(ProcessorResponse)
	return provisionerResponse.Success
}

func GetClusters(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	moduleName string, maxWaitForResponse float64) (clusters []processor.ClusterData, success bool) {

	request := ProcessorRequest{
		Action:      ProcessorClusterGet,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: "", Config: ""},
		Source:      HttpClient,
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(ProcessorResponse)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).([]processor.ClusterData), true
}

func CreateSupervisor(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	moduleName, clusterName, configName string, metadata map[string]string, maxWaitForResponse float64) (uint64, error) {

	request := ProcessorRequest{
		Action:      ProcessorSupervisorCreate,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: configName},
		Data:        metadata,
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)

	if didTimeout {
		return 0, multithreaded.NoResponseReceived
	}

	response := (rsp).(ProcessorResponse)

	return (response.Data).(uint64), response.Error
}

func GetSupervisor(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable, timeout float64,
	id uint64) (*supervisor.Supervisor, error) {

	request := ProcessorRequest{
		Action:      ProcessorSupervisorGet,
		Identifiers: RequestIdentifiers{Supervisor: id},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (data).(ProcessorResponse)

	if !response.Success {
		return nil, response.Error
	}

	return (response.Data).(*supervisor.Supervisor), nil
}

func UpdateSupervisor(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable, timeout float64,
	data *supervisor.Supervisor) error {

	request := ProcessorRequest{
		Action: ProcessorSupervisorUpdate,
		Data:   data,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(ProcessorResponse)
	return response.Error
}

func FindStatistics(pipe chan<- DatabaseRequest, responseTable *multithreaded.ResponseTable,
	moduleName, clusterName string, maxWaitForResponse float64) (entries []database.Statistic, found bool) {

	databaseRequest := DatabaseRequest{
		Action:  DatabaseFetch,
		Type:    SupervisorStatistic,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(responseTable, databaseRequest.Nonce, maxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(DatabaseResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	return (databaseResponse.Data).([]database.Statistic), true
}

func ShutdownCore(pipe chan<- InterruptEvent) (response []byte, success bool) {
	pipe <- Shutdown
	return nil, true
}

// TODO : this needs to be fixed
//func PingNodeChannels(logger *multithreaded.Logger,
//	databasePipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
//	provisionerPipe chan<- ProvisionerRequest, provisionerResponseTable *multithreaded.ResponseTable) (success bool) {
//
//	databasePingRequest := DatabaseRequest{
//		Action: DatabaseUpperPing,
//		Nonce:  rand.Uint32(),
//	}
//	databasePipe <- databasePingRequest
//
//	data, didTimeout := multithreaded.SendAndWait(databaseResponseTable, databasePingRequest.Nonce,
//		maxWaitForResponse)
//	if didTimeout {
//		return false
//	}
//
//	databaseResponse := (data).(DatabaseResponse)
//	if !databaseResponse.Success {
//		return false
//	}
//
//	if GetConfigInstance().Debug {
//		logger.Println("received ping over C2")
//	}
//
//	provisionerPingRequest := ProvisionerRequest{
//		Action: ProvisionerLowerPing,
//		Source: Http,
//		Nonce:  rand.Uint32(),
//	}
//	provisionerPipe <- provisionerPingRequest
//
//	data2, didTimeout2 := multithreaded.SendAndWait(provisionerResponseTable, provisionerPingRequest.Nonce,
//		maxWaitForResponse)
//	if didTimeout2 {
//		return false
//	}
//
//	provisionerResponse := (data2).(ProvisionerResponse)
//	if !provisionerResponse.Success {
//		return false
//	}
//
//	if GetConfigInstance().Debug {
//		logger.Println("received ping over C6")
//	}
//
//	return true
//}

func GetModules(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable, maxWaitForResponse float64) (success bool, modules []processor.ModuleData) {

	request := ProcessorRequest{
		Action: ProcessorModuleGet,
		Source: HttpClient,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return false, nil
	}

	provisionerResponse := (data).(ProcessorResponse)

	if !provisionerResponse.Success {
		return false, nil
	}

	return true, (provisionerResponse.Data).([]processor.ModuleData)
}

func AddModule(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	processorName string, cfg *module.Config, maxWaitForResponse float64) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleAdd,
		Source:      HttpProcessor,
		Identifiers: RequestIdentifiers{Processor: processorName},
		Data:        *cfg,
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

func MountModule(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	moduleName string, maxWaitForResponse float64) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleMount,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

func UnmountModule(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	moduleName string, maxWaitForResponse float64) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleUnmount,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

func DeleteModule(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable,
	host string, port int, moduleName string, maxWaitForResponse float64) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleRemove,
		Source:      HttpProcessor,
		Identifiers: RequestIdentifiers{Processor: fmt.Sprintf("%s:%d", host, port), Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)

	if didTimeout {
		return false, multithreaded.NoResponseReceived
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

// TODO : I believe this needs to be removed from the core
//func RegisterModule(pipe chan<- ProvisionerRequest, responseTable *multithreaded.ResponseTable,
//	modulePath string) (success bool, description string) {
//
//	request := ProvisionerRequest{
//		Action: ProvisionerModuleLoad,
//		Source: Http,
//		Metadata: ProvisionerMetadata{
//			ModulePath: modulePath,
//		},
//		Nonce: rand.Uint32(),
//	}
//	pipe <- request
//
//	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce,
//		maxWaitForResponse)
//	if didTimeout {
//		return false, "timeout"
//	}
//
//	provisionerResponse := (data).(ProvisionerResponse)
//	return provisionerResponse.Success, provisionerResponse.Description
//}

// TODO : I believe this needs to be removed from the core
//func DeleteModule(pipe chan<- ProvisionerRequest, responseTable *multithreaded.ResponseTable,
//	moduleName string) (success bool, description string) {
//
//	request := ProvisionerRequest{
//		Action:     ProvisionerModuleDelete,
//		Source:     Http,
//		ModuleName: moduleName,
//		Nonce:      rand.Uint32(),
//	}
//	pipe <- request
//
//	data, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce,
//		maxWaitForResponse)
//	if didTimeout {
//		return false, "timeout"
//	}
//
//	provisionerResponse := (data).(ProvisionerResponse)
//
//	return provisionerResponse.Success, provisionerResponse.Description
//}

func FetchFromCache(pipe chan<- CacheRequest, responseTable *multithreaded.ResponseTable,
	key string, maxWaitForResponse float64) (value any, found bool) {

	request := CacheRequest{
		Action:     CacheLoadFrom,
		Identifier: key,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)

	if didTimeout {
		return nil, false
	}

	response := (rsp).(CacheResponse)

	return response.Data, response.Success
}

func StoreInCache(pipe chan<- CacheRequest, responseTable *multithreaded.ResponseTable,
	data any, expiry float64, maxWaitForResponse float64) (identifier string, success bool) {

	request := CacheRequest{
		Action:    CacheSaveIn,
		Data:      data,
		Nonce:     rand.Uint32(),
		ExpiresIn: expiry,
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)

	if didTimeout {
		success = false
	} else {

	}

	response := (rsp).(CacheResponse)
	return response.Identifier, response.Success
}

func SwapInCache(pipe chan<- CacheRequest, responseTable *multithreaded.ResponseTable,
	key string, data any, maxWaitForResponse float64) (success bool) {

	request := CacheRequest{
		Action:     CacheSaveIn,
		Data:       data,
		Identifier: key,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, maxWaitForResponse)

	if didTimeout {
		return false
	}

	response := (rsp).(CacheResponse)
	return response.Success
}

func Log(pipe chan<- ProcessorRequest, responseTable *multithreaded.ResponseTable, timeout float64,
	log *supervisor.Log) error {

	if log == nil {
		return errors.New("need a valid *supervisor.Log")
	}

	request := ProcessorRequest{
		Action: ProcessorSupervisorLog,
		Data:   log,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(ProcessorResponse)
	return response.Error
}
