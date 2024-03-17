package common

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/core/components/database"
	"github.com/GabeCordo/cluster-tools/core/components/processor"
	"github.com/GabeCordo/cluster-tools/core/components/scheduler"
	"github.com/GabeCordo/cluster-tools/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

func GetConfigFromDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName, clusterName string, maxWaitForResponse float64) (conf interfaces.Config, found bool) {

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
		return interfaces.Config{}, false
	}

	databaseResponse := (data).(DatabaseResponse)

	if !databaseResponse.Success {
		return interfaces.Config{}, false
	}

	configs := (databaseResponse.Data).([]interfaces.Config)
	if len(configs) != 1 {
		return interfaces.Config{}, false
	}

	return configs[0], true
}

func GetConfigsFromDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName string, maxWaitForResponse float64) (configs []interfaces.Config, found bool) {

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

	configs = (databaseResponse.Data).([]interfaces.Config)
	return configs, true
}

func StoreConfigInDatabase(pipe chan<- DatabaseRequest, databaseResponseTable *multithreaded.ResponseTable,
	moduleName string, cfg interfaces.Config, maxWaitForResponse float64) error {

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
	moduleName string, cfg interfaces.Config, maxWaitForResponse float64) (success bool) {

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
	cfg *interfaces.ProcessorConfig, maxWaitForResponse float64) (bool, error) {

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
	cfg *interfaces.ProcessorConfig, maxWaitForResponse float64) error {

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
	filter supervisor.Filter) ([]*supervisor.Supervisor, error) {

	request := ProcessorRequest{
		Action: ProcessorSupervisorGet,
		Identifiers: RequestIdentifiers{
			Module:     filter.Module,
			Cluster:    filter.Cluster,
			Supervisor: filter.Id,
		},
		Nonce: rand.Uint32(),
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

	return (response.Data).([]*supervisor.Supervisor), nil
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

func ShutdownCore(pipe chan<- InterruptEvent) error {
	pipe <- Shutdown
	return nil
}

func PingNodeChannels(databasePipe chan<- DatabaseRequest, databaseRT *multithreaded.ResponseTable,
	processorPipe chan<- ProcessorRequest, processorRT *multithreaded.ResponseTable, timeout float64) error {

	databasePingRequest := DatabaseRequest{
		Action: DatabaseUpperPing,
		Nonce:  rand.Uint32(),
	}
	databasePipe <- databasePingRequest

	data, didTimeout := multithreaded.SendAndWait(databaseRT, databasePingRequest.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	databaseResponse := (data).(DatabaseResponse)
	if !databaseResponse.Success {
		return databaseResponse.Error
	}

	fmt.Println("received ping over C2")

	processorRequest := ProcessorRequest{
		Action: ProcessorPing,
		Nonce:  rand.Uint32(),
	}
	processorPipe <- processorRequest

	data2, didTimeout2 := multithreaded.SendAndWait(processorRT, processorRequest.Nonce, timeout)
	if didTimeout2 {
		return multithreaded.NoResponseReceived
	}

	processorResponse := (data2).(ProcessorResponse)
	if !processorResponse.Success {
		return processorResponse.Error
	}

	fmt.Println("received ping over C6")

	return nil
}

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
	processorName string, cfg *interfaces.ModuleConfig, maxWaitForResponse float64) (bool, error) {

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
//func AddModule(pipe chan<- ProvisionerRequest, responseTable *multithreaded.ResponseTable,
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

	//HOTFIX : too long to response to the log request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(ProcessorResponse)
	return response.Error
	//return nil
}

func GetJobs(pipe chan<- SchedulerRequest, responseTable *multithreaded.ResponseTable, timeout float64,
	filter *scheduler.Filter) ([]scheduler.Job, error) {

	request := SchedulerRequest{
		Action: SchedulerGet,
		Data:   *filter,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(SchedulerResponse)

	return (response.Data).([]scheduler.Job), nil
}

func CreateJob(pipe chan<- SchedulerRequest, responseTable *multithreaded.ResponseTable, timeout float64,
	job *scheduler.Job) error {

	request := SchedulerRequest{
		Action: SchedulerCreate,
		Data:   *job,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(SchedulerResponse)
	return response.Error
}

func DeleteJob(pipe chan<- SchedulerRequest, responseTable *multithreaded.ResponseTable, timeout float64,
	filter *scheduler.Filter) error {

	request := SchedulerRequest{
		Action: SchedulerDelete,
		Data:   *filter,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(SchedulerResponse)
	return response.Error
}

func JobQueue(pipe chan<- SchedulerRequest, responseTable *multithreaded.ResponseTable, timeout float64) ([]scheduler.Job, error) {

	request := SchedulerRequest{
		Action: SchedulerGet,
		Type:   SchedulerQueue,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(SchedulerResponse)
	return (response.Data).([]scheduler.Job), response.Error
}

func GetSubscribers(pipe chan<- MessengerRequest, responseTable *multithreaded.ResponseTable, timeout float64) ([]string, error) {

	request := MessengerRequest{
		Action: MessengerGetSubscribers,
		Nonce:  rand.Uint32(),
	}
	pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(responseTable, request.Nonce, timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response, ok := (rsp).(MessengerResponse)
	if !ok {
		return nil, InternalError
	}

	subscribers, ok := (response.Data).([]string)
	if !ok {
		return nil, InternalError
	}

	return subscribers, response.Error
}
