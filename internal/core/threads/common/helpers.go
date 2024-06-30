package common

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/components/database"
	"github.com/GabeCordo/cluster-tools/internal/core/components/processor"
	"github.com/GabeCordo/cluster-tools/internal/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
)

type ThreadMandatory struct {
	Pipe          chan<- ThreadRequest
	ResponseTable *multithreaded.ResponseTable
	Timeout       float64
}

func GetConfigFromDatabase(mandatory ThreadMandatory, moduleName, clusterName string) (conf interfaces.Config, found bool) {

	databaseRequest := ThreadRequest{
		Action: GetAction,
		Type:   ConfigRecord,
		Identifiers: RequestIdentifiers{
			Module:  moduleName,
			Cluster: clusterName,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return interfaces.Config{}, false
	}

	databaseResponse := (data).(ThreadResponse)

	if !databaseResponse.Success {
		return interfaces.Config{}, false
	}

	configs := (databaseResponse.Data).([]interfaces.Config)
	if len(configs) != 1 {
		return interfaces.Config{}, false
	}

	return configs[0], true
}

func GetConfigsFromDatabase(mandatory ThreadMandatory, moduleName string) (configs []interfaces.Config, found bool) {

	databaseRequest := ThreadRequest{
		Action:      GetAction,
		Type:        ConfigRecord,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(ThreadResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	configs = (databaseResponse.Data).([]interfaces.Config)
	return configs, true
}

func StoreConfigInDatabase(mandatory ThreadMandatory, moduleName string, cfg interfaces.Config) error {

	databaseRequest := ThreadRequest{
		Action: CreateAction,
		Type:   ConfigRecord,
		Identifiers: RequestIdentifiers{
			Module:  moduleName,
			Cluster: cfg.Identifier,
		},
		Data:  cfg,
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	databaseResponse := (data).(ThreadResponse)
	// TODO : make the database generate the errors
	if !databaseResponse.Success {
		return errors.New("could not store config in database")
	}

	return nil
}

func ReplaceConfigInDatabase(mandatory ThreadMandatory, moduleName string, cfg interfaces.Config) (success bool) {

	databaseRequest := ThreadRequest{
		Action: UpdateAction,
		Type:   ConfigRecord,
		Identifiers: RequestIdentifiers{
			Module:  moduleName,
			Cluster: cfg.Identifier,
		},
		Data:  cfg,
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(ThreadResponse)
	return databaseResponse.Success
}

func DeleteConfigInDatabase(mandatory ThreadMandatory, moduleName, configName string) (success bool) {

	databaseRequest := ThreadRequest{
		Action: DeleteAction,
		Type:   ConfigRecord,
		Identifiers: RequestIdentifiers{
			Module:  moduleName,
			Cluster: configName,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(ThreadResponse)
	return databaseResponse.Success
}

func GetProcessors(mandatory ThreadMandatory) ([]*processor.Processor, bool) {

	request := ThreadRequest{
		Action: GetAction,
		Type:   ProcessorRecord,
		Source: HttpClient,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	response := (data).(ThreadResponse)

	if response.Success {
		return (response.Data).([]*processor.Processor), true
	} else {
		return nil, false
	}
}

func AddProcessor(mandatory ThreadMandatory, cfg *interfaces.ProcessorConfig) (bool, error) {

	request := ThreadRequest{
		Action: CreateAction,
		Type:   ProcessorRecord,
		Source: HttpProcessor,
		Data:   *cfg,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ThreadResponse)
	return response.Success, response.Error
}

func DeleteProcessor(mandatory ThreadMandatory, cfg *interfaces.ProcessorConfig) error {

	request := ThreadRequest{
		Action: DeleteAction,
		Type:   ProcessorRecord,
		Source: HttpProcessor,
		Data:   *cfg,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (data).(ThreadResponse)
	return response.Error
}

func MountCluster(mandatory ThreadMandatory, moduleName, clusterName string) (success bool) {

	request := ThreadRequest{
		Action:      MountAction,
		Type:        ClusterRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(ThreadResponse)
	return provisionerResponse.Success
}

func UnmountCluster(mandatory ThreadMandatory, moduleName, clusterName string) (success bool) {

	request := ThreadRequest{
		Action:      UnMountAction,
		Type:        ClusterRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(ThreadResponse)
	return provisionerResponse.Success
}

func GetClusters(mandatory ThreadMandatory, moduleName string) (clusters []processor.ClusterData, success bool) {

	request := ThreadRequest{
		Action:      GetAction,
		Type:        ClusterRecord,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: "", Config: ""},
		Source:      HttpClient,
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(ThreadResponse)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).([]processor.ClusterData), true
}

func CreateSupervisor(mandatory ThreadMandatory,
	moduleName, clusterName, configName string, metadata map[string]string) (uint64, error) {

	request := ThreadRequest{
		Action:      CreateAction,
		Type:        SupervisorRecord,
		Identifiers: RequestIdentifiers{Module: moduleName, Cluster: clusterName, Config: configName},
		Data:        metadata,
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return 0, multithreaded.NoResponseReceived
	}

	response := (rsp).(ThreadResponse)

	return (response.Data).(uint64), response.Error
}

func GetSupervisor(mandatory ThreadMandatory, filter supervisor.Filter) ([]*supervisor.Supervisor, error) {

	request := ThreadRequest{
		Action: GetAction,
		Type:   SupervisorRecord,
		Identifiers: RequestIdentifiers{
			Module:     filter.Module,
			Cluster:    filter.Cluster,
			Supervisor: filter.Id,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (data).(ThreadResponse)

	if !response.Success {
		return nil, response.Error
	}

	return (response.Data).([]*supervisor.Supervisor), nil
}

func UpdateSupervisor(mandatory ThreadMandatory, data *supervisor.Supervisor) error {

	request := ThreadRequest{
		Action: UpdateAction,
		Type:   SupervisorRecord,
		Data:   data,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(ThreadResponse)
	return response.Error
}

func FindStatistics(mandatory ThreadMandatory, moduleName, clusterName string) (entries []database.Statistic, found bool) {

	databaseRequest := ThreadRequest{
		Action: GetAction,
		Type:   StatisticRecord,
		Identifiers: RequestIdentifiers{
			Module:  moduleName,
			Cluster: clusterName,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(ThreadResponse)

	if !databaseResponse.Success {
		return nil, false
	}

	return (databaseResponse.Data).([]database.Statistic), true
}

func ShutdownCore(pipe chan<- InterruptEvent) error {
	pipe <- Shutdown
	return nil
}

func GetModules(mandatory ThreadMandatory) (success bool, modules []processor.ModuleData) {

	request := ThreadRequest{
		Action: GetAction,
		Type:   ModuleRecord,
		Source: HttpClient,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, nil
	}

	provisionerResponse := (data).(ThreadResponse)

	if !provisionerResponse.Success {
		return false, nil
	}

	return true, (provisionerResponse.Data).([]processor.ModuleData)
}

func AddModule(mandatory ThreadMandatory, processorName string, cfg *interfaces.ModuleConfig) (bool, error) {

	request := ThreadRequest{
		Action:      CreateAction,
		Type:        ModuleRecord,
		Source:      HttpProcessor,
		Identifiers: RequestIdentifiers{Processor: processorName},
		Data:        *cfg,
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ThreadResponse)

	return response.Success, response.Error
}

func MountModule(mandatory ThreadMandatory, moduleName string) (bool, error) {

	request := ThreadRequest{
		Action:      MountAction,
		Type:        ModuleRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ThreadResponse)

	return response.Success, response.Error
}

func UnmountModule(mandatory ThreadMandatory, moduleName string) (bool, error) {

	request := ThreadRequest{
		Action:      UnMountAction,
		Type:        ModuleRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(ThreadResponse)

	return response.Success, response.Error
}

func DeleteModule(mandatory ThreadMandatory, host string, port int, moduleName string) (bool, error) {

	request := ThreadRequest{
		Action:      DeleteAction,
		Type:        ModuleRecord,
		Source:      HttpProcessor,
		Identifiers: RequestIdentifiers{Processor: fmt.Sprintf("%s:%d", host, port), Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return false, multithreaded.NoResponseReceived
	}

	response := (data).(ThreadResponse)

	return response.Success, response.Error
}

func FetchFromCache(mandatory ThreadMandatory, key string) (value any, found bool) {

	request := ThreadRequest{
		Action: GetAction,
		Type:   CacheRecord,
		Data: CacheRequestData{
			Identifier: key,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return nil, false
	}

	response := (rsp).(ThreadResponse)

	return response.Data, response.Success
}

func StoreInCache(mandatory ThreadMandatory, data any, expiry float64) (identifier string, success bool) {

	request := ThreadRequest{
		Action: CreateAction,
		Type:   CacheRecord,
		Data: CacheRequestData{
			ExpiresIn: expiry,
			Data:      data,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		success = false
	} else {

	}

	response := (rsp).(ThreadResponse)

	cacheResponseData := (response.Data).(CacheResponseData)

	return cacheResponseData.Identifier, response.Success
}

func SwapInCache(mandatory ThreadMandatory, key string, data any) (success bool) {

	request := ThreadRequest{
		Action: CreateAction,
		Type:   CacheRecord,
		Data: CacheRequestData{
			Identifier: key,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return false
	}

	response := (rsp).(ThreadResponse)
	return response.Success
}

func Log(mandatory ThreadMandatory, log *supervisor.Log) error {

	if log == nil {
		return errors.New("need a valid *supervisor.Log")
	}

	request := ThreadRequest{
		Action: LogAction,
		Type:   SupervisorRecord,
		Data:   log,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	//HOTFIX : too long to response to the log request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(ThreadResponse)
	return response.Error
	//return nil
}

func GetJobs(mandatory ThreadMandatory, filter *interfaces.Filter) ([]interfaces.Job, error) {

	request := ThreadRequest{
		Action: GetAction,
		Type:   JobRecord,
		Data:   *filter,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(ThreadResponse)

	return (response.Data).([]interfaces.Job), nil
}

func CreateJob(mandatory ThreadMandatory, job *interfaces.Job) error {

	request := ThreadRequest{
		Action: CreateAction,
		Type:   JobRecord,
		Data:   *job,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(ThreadResponse)
	return response.Error
}

func DeleteJob(mandatory ThreadMandatory, filter *interfaces.Filter) error {

	request := ThreadRequest{
		Action: DeleteAction,
		Type:   JobRecord,
		Data:   *filter,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return multithreaded.NoResponseReceived
	}

	response := (rsp).(ThreadResponse)
	return response.Error
}

func JobQueue(mandatory ThreadMandatory) ([]interfaces.Job, error) {

	request := ThreadRequest{
		Action: GetAction,
		Type:   QueueRecord,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(ThreadResponse)
	return (response.Data).([]interfaces.Job), response.Error
}

func GetSubscribers(mandatory ThreadMandatory) ([]string, error) {

	request := ThreadRequest{
		Action: GetAction,
		Type:   SubscriberRecord,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response, ok := (rsp).(ThreadResponse)
	if !ok {
		return nil, InternalError
	}

	subscribers, ok := (response.Data).([]string)
	if !ok {
		return nil, InternalError
	}

	return subscribers, response.Error
}
