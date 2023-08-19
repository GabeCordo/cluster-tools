package common

import (
	"errors"
	"github.com/GabeCordo/mango-core/core/components/database"
	"github.com/GabeCordo/mango-core/core/components/processor"
	"github.com/GabeCordo/mango/components/cluster"
	"github.com/GabeCordo/mango/module"
	processor_i "github.com/GabeCordo/mango/processor"
	"github.com/GabeCordo/mango/threads"
	"github.com/GabeCordo/mango/utils"
	"math/rand"
)

func GetConfigFromDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName, clusterName string) (conf cluster.Config, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseFetch,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return cluster.Config{}, false
	}

	databaseResponse := (data).(threads.DatabaseResponse)

	if !databaseResponse.Success {
		return cluster.Config{}, false
	}

	configs := (databaseResponse.Data).([]cluster.Config)
	if len(configs) != 1 {
		return cluster.Config{}, false
	}

	return configs[0], true
}

func GetConfigsFromDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName string) (configs []cluster.Config, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action: threads.DatabaseFetch,
		Type:   threads.ClusterConfig,
		Module: moduleName,
		Nonce:  rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(threads.DatabaseResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	configs = (databaseResponse.Data).([]cluster.Config)
	return configs, true
}

func StoreConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName string, cfg cluster.Config) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseStore,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: cfg.Identifier,
		Data:    cfg,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(threads.DatabaseResponse)
	return databaseResponse.Success
}

func ReplaceConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName string, cfg cluster.Config) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseReplace,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: cfg.Identifier,
		Data:    cfg,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(threads.DatabaseResponse)
	return databaseResponse.Success
}

func DeleteConfigInDatabase(pipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
	moduleName, configName string) (success bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseDelete,
		Type:    threads.ClusterConfig,
		Module:  moduleName,
		Cluster: configName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(databaseResponseTable, databaseRequest.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(threads.DatabaseResponse)
	return databaseResponse.Success
}

func GetProcessors(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable) ([]processor.Processor, bool) {

	request := ProcessorRequest{
		Action: ProcessorGet,
		Source: threads.HttpClient,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
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

func AddProcessor(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	cfg *processor_i.Config) (bool, error) {

	request := ProcessorRequest{
		Action: ProcessorAdd,
		Source: threads.HttpProcessor,
		Data:   *cfg,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)
	return response.Success, response.Error
}

func DeleteProcessor(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	processorName string) error {

	request := ProcessorRequest{
		Action:      ProcessorRemove,
		Source:      threads.HttpProcessor,
		Identifiers: RequestIdentifiers{Processor: processorName},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return utils.NoResponseReceived
	}

	response := (data).(ProcessorResponse)
	return response.Error
}

func MountCluster(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string) (success bool) {

	request := ProcessorRequest{
		Action:      ProcessorClusterMount,
		Source:      threads.HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(ProcessorResponse)
	return provisionerResponse.Success
}

func UnmountCluster(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName string) (success bool) {

	request := ProcessorRequest{
		Action:      ProcessorClusterUnmount,
		Source:      threads.HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(ProcessorResponse)
	return provisionerResponse.Success
}

func GetClusters(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	moduleName string) (clusters []processor.ClusterData, success bool) {

	request := ProcessorRequest{
		Action:      ProcessorClusterGet,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: "", Config: ""},
		Source:      threads.HttpClient,
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(ProcessorResponse)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).([]processor.ClusterData), true
}

func CreateSupervisor(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	moduleName, clusterName, configName string, metadata map[string]string) (uint64, error) {

	request := ProcessorRequest{
		Action:      ProcessorSupervisorCreate,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: configName},
		Data:        metadata,
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		return 0, utils.NoResponseReceived
	}

	response := (rsp).(ProcessorResponse)

	return (response.Data).(uint64), response.Error
}

// TODO : fix
//func GetSupervisors(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
//	moduleName, clusterName string) (map[uint64]supervisor.Status, bool) {
//
//	request := threads.ProvisionerRequest{
//		Action:      threads.ProvisionerGetSupervisors,
//		Source:      threads.Http,
//		ModuleName:  moduleName,
//		ClusterName: clusterName,
//		Nonce:       rand.Uint32(),
//	}
//	pipe <- request
//
//	data, timeout := utils.SendAndWait(responseTable, request.Nonce, GetConfigInstance().MaxWaitForResponse)
//	if timeout {
//		return nil, false
//	}
//	provisionerResponse := (data).(threads.ProvisionerResponse)
//
//	return (provisionerResponse.data).(map[uint64]supervisor.Status), true
//}

// TODO : fix
//func GetSupervisor(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
//	moduleName, clusterName string, supervisorId uint64) (supervisorInstance *supervisor.Supervisor, success bool) {
//
//	request := threads.ProvisionerRequest{
//		Action:      threads.ProvisionerGetSupervisor,
//		Source:      threads.Http,
//		ModuleName:  moduleName,
//		ClusterName: clusterName,
//		Metadata: threads.ProvisionerMetadata{
//			SupervisorId: supervisorId,
//		},
//		Nonce: rand.Uint32(),
//	}
//	pipe <- request
//
//	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce, GetConfigInstance().MaxWaitForResponse)
//	if didTimeout {
//		return nil, false
//	}
//
//	provisionerResponse := (data).(threads.ProvisionerResponse)
//
//	if !provisionerResponse.Success {
//		return nil, false
//	}
//
//	return (provisionerResponse.data).(*supervisor.Supervisor), true
//}

func FindStatistics(pipe chan<- threads.DatabaseRequest, responseTable *utils.ResponseTable, moduleName, clusterName string) (entries []database.Statistic, found bool) {

	databaseRequest := threads.DatabaseRequest{
		Action:  threads.DatabaseFetch,
		Type:    threads.SupervisorStatistic,
		Module:  moduleName,
		Cluster: clusterName,
		Nonce:   rand.Uint32(),
	}
	pipe <- databaseRequest

	data, didTimeout := utils.SendAndWait(responseTable, databaseRequest.Nonce, GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(threads.DatabaseResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	return (databaseResponse.Data).([]database.Statistic), true
}

func ShutdownCore(pipe chan<- threads.InterruptEvent) (response []byte, success bool) {
	pipe <- threads.Shutdown
	return nil, true
}

// TODO : this needs to be fixed
//func PingNodeChannels(logger *utils.Logger,
//	databasePipe chan<- threads.DatabaseRequest, databaseResponseTable *utils.ResponseTable,
//	provisionerPipe chan<- threads.ProvisionerRequest, provisionerResponseTable *utils.ResponseTable) (success bool) {
//
//	databasePingRequest := threads.DatabaseRequest{
//		Action: threads.DatabaseUpperPing,
//		Nonce:  rand.Uint32(),
//	}
//	databasePipe <- databasePingRequest
//
//	data, didTimeout := utils.SendAndWait(databaseResponseTable, databasePingRequest.Nonce,
//		GetConfigInstance().MaxWaitForResponse)
//	if didTimeout {
//		return false
//	}
//
//	databaseResponse := (data).(threads.DatabaseResponse)
//	if !databaseResponse.Success {
//		return false
//	}
//
//	if GetConfigInstance().Debug {
//		logger.Println("received ping over C2")
//	}
//
//	provisionerPingRequest := threads.ProvisionerRequest{
//		Action: threads.ProvisionerLowerPing,
//		Source: threads.Http,
//		Nonce:  rand.Uint32(),
//	}
//	provisionerPipe <- provisionerPingRequest
//
//	data2, didTimeout2 := utils.SendAndWait(provisionerResponseTable, provisionerPingRequest.Nonce,
//		GetConfigInstance().MaxWaitForResponse)
//	if didTimeout2 {
//		return false
//	}
//
//	provisionerResponse := (data2).(threads.ProvisionerResponse)
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

func GetModules(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable) (success bool, modules []processor.ModuleData) {

	request := ProcessorRequest{
		Action: ProcessorModuleGet,
		Source: threads.HttpClient,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, nil
	}

	provisionerResponse := (data).(ProcessorResponse)

	if !provisionerResponse.Success {
		return false, nil
	}

	return true, (provisionerResponse.Data).([]processor.ModuleData)
}

func AddModule(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	processorName string, cfg *module.Config) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleAdd,
		Source:      threads.HttpProcessor,
		Identifiers: RequestIdentifiers{Processor: processorName},
		Data:        *cfg,
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

func MountModule(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	moduleName string) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleMount,
		Source:      threads.HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

func UnmountModule(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	moduleName string) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleUnmount,
		Source:      threads.HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

func DeleteModule(pipe chan<- ProcessorRequest, responseTable *utils.ResponseTable,
	processorName, moduleName string) (bool, error) {

	request := ProcessorRequest{
		Action:      ProcessorModuleDelete,
		Source:      threads.HttpProcessor,
		Identifiers: RequestIdentifiers{Processor: processorName, Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	pipe <- request

	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		return false, utils.NoResponseReceived
	}

	response := (data).(ProcessorResponse)

	return response.Success, response.Error
}

// TODO : I believe this needs to be removed from the core
//func RegisterModule(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
//	modulePath string) (success bool, description string) {
//
//	request := threads.ProvisionerRequest{
//		Action: threads.ProvisionerModuleLoad,
//		Source: threads.Http,
//		Metadata: threads.ProvisionerMetadata{
//			ModulePath: modulePath,
//		},
//		Nonce: rand.Uint32(),
//	}
//	pipe <- request
//
//	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
//		GetConfigInstance().MaxWaitForResponse)
//	if didTimeout {
//		return false, "timeout"
//	}
//
//	provisionerResponse := (data).(threads.ProvisionerResponse)
//	return provisionerResponse.Success, provisionerResponse.Description
//}

// TODO : I believe this needs to be removed from the core
//func DeleteModule(pipe chan<- threads.ProvisionerRequest, responseTable *utils.ResponseTable,
//	moduleName string) (success bool, description string) {
//
//	request := threads.ProvisionerRequest{
//		Action:     threads.ProvisionerModuleDelete,
//		Source:     threads.Http,
//		ModuleName: moduleName,
//		Nonce:      rand.Uint32(),
//	}
//	pipe <- request
//
//	data, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
//		GetConfigInstance().MaxWaitForResponse)
//	if didTimeout {
//		return false, "timeout"
//	}
//
//	provisionerResponse := (data).(threads.ProvisionerResponse)
//
//	return provisionerResponse.Success, provisionerResponse.Description
//}

func FetchFromCache(pipe chan<- threads.CacheRequest, responseTable *utils.ResponseTable,
	key string) (value any, found bool) {

	request := threads.CacheRequest{
		Action:     threads.CacheLoadFrom,
		Identifier: key,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		return nil, false
	}

	response := (rsp).(threads.CacheResponse)

	return response.Data, response.Success
}

func StoreInCache(pipe chan<- threads.CacheRequest, responseTable *utils.ResponseTable,
	data any, expiry float64) (identifier string, success bool) {

	request := threads.CacheRequest{
		Action:    threads.CacheSaveIn,
		Data:      data,
		Nonce:     rand.Uint32(),
		ExpiresIn: expiry,
	}
	pipe <- request

	rsp, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		success = false
	} else {

	}

	response := (rsp).(threads.CacheResponse)
	return response.Identifier, response.Success
}

func SwapInCache(pipe chan<- threads.CacheRequest, responseTable *utils.ResponseTable,
	key string, data any) (success bool) {

	request := threads.CacheRequest{
		Action:     threads.CacheSaveIn,
		Data:       data,
		Identifier: key,
		Nonce:      rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := utils.SendAndWait(responseTable, request.Nonce,
		GetConfigInstance().MaxWaitForResponse)

	if didTimeout {
		return false
	}

	response := (rsp).(threads.CacheResponse)
	return response.Success
}

func ToggleDebugMode(logger *utils.Logger) (description string) {

	cfg := GetConfigInstance()
	cfg.Debug = !cfg.Debug

	if cfg.Debug {
		description = "debug mode activated"
		logger.Println("remote change: debug mode ON")
	} else {
		description = "debug mode disabled"
		logger.Println("remote change: debug mode OFF")
	}

	return description
}
