package thread

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database"
	"github.com/GabeCordo/cluster-tools/internal/core/database/job"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/core/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/core/database/supervisor"
	"github.com/GabeCordo/cluster-tools/internal/core/message/log"
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"github.com/GabeCordo/toolchain/multithreaded"
	"math/rand"
	"strconv"
)

type Mandatory struct {
	Pipe          chan<- Request
	ResponseTable *multithreaded.ResponseTable
	Timeout       float64
}

func GetPipelineFromDatabase(mandatory Mandatory, moduleName, clusterName string) (conf pipeline.Pipeline, found bool) {

	databaseRequest := Request{
		Action: GetAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Module:   moduleName,
			Function: clusterName,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return pipeline.Pipeline{}, false
	}

	databaseResponse := (data).(Response)

	if !databaseResponse.Success {
		return pipeline.Pipeline{}, false
	}
	return databaseResponse.Data.([]pipeline.Pipeline)[0], true
}

func GetPipelinesFromDatabase(mandatory Mandatory, moduleName string) (configs []pipeline.Pipeline, found bool) {

	databaseRequest := Request{
		Action:      GetAction,
		Type:        PipelineRecord,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(Response)

	if !databaseResponse.Success {
		return nil, false
	}
	return databaseResponse.Data.([]pipeline.Pipeline), true
}

func StorePipelineInDatabase(mandatory Mandatory, moduleName string, cfg pipeline.Pipeline) error {

	databaseRequest := Request{
		Action: CreateAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Module:   moduleName,
			Function: cfg.Identifier,
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

	databaseResponse := (data).(Response)
	// TODO : make the database generate the errors
	if !databaseResponse.Success {
		return errors.New("could not database pipeline in database")
	}

	return nil
}

func ReplacePipelineInDatabase(mandatory Mandatory, moduleName string, cfg pipeline.Pipeline) (success bool) {

	databaseRequest := Request{
		Action: UpdateAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Module:   moduleName,
			Function: cfg.Identifier,
		},
		Data:  cfg,
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(Response)
	return databaseResponse.Success
}

func DeletePipelineInDatabase(mandatory Mandatory, moduleName, configName string) (success bool) {

	databaseRequest := Request{
		Action: DeleteAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Module: moduleName,
			Config: configName,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(Response)
	return databaseResponse.Success
}

func GetProcessors(mandatory Mandatory) ([]*processor.Processor, bool) {

	request := Request{
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

	response := (data).(Response)

	if response.Success {
		return (response.Data).([]*processor.Processor), true
	} else {
		return nil, false
	}
}

func AddProcessor(mandatory Mandatory, cfg *processor.Config) (bool, error) {

	request := Request{
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

	response := (data).(Response)
	return response.Success, response.Error
}

func DeleteProcessor(mandatory Mandatory, cfg *processor.Config) error {

	request := Request{
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

	response := (data).(Response)
	return response.Error
}

func MountCluster(mandatory Mandatory, moduleName, clusterName string) (success bool) {

	request := Request{
		Action:      MountAction,
		Type:        FunctionRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Function: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(Response)
	return provisionerResponse.Success
}

func UnmountCluster(mandatory Mandatory, moduleName, clusterName string) (success bool) {

	request := Request{
		Action:      UnMountAction,
		Type:        FunctionRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Function: clusterName, Config: ""},
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(Response)
	return provisionerResponse.Success
}

func GetFunctions(mandatory Mandatory, moduleName string) (clusters []processor.FunctionData, success bool) {

	request := Request{
		Action:      GetAction,
		Type:        FunctionRecord,
		Identifiers: RequestIdentifiers{Module: moduleName, Function: "", Config: ""},
		Source:      HttpClient,
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(Response)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).([]processor.FunctionData), true
}

func CreateSupervisor(mandatory Mandatory,
	moduleName, clusterName, configName string, metadata map[string]string) (uint64, error) {

	request := Request{
		Action:      CreateAction,
		Type:        SupervisorRecord,
		Identifiers: RequestIdentifiers{Module: moduleName, Function: clusterName, Config: configName},
		Data:        metadata,
		Nonce:       rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return 0, multithreaded.NoResponseReceived
	}

	response := (rsp).(Response)

	return (response.Data).(uint64), response.Error
}

func GetSupervisor(mandatory Mandatory, filter database.Filter) ([]*supervisor.Supervisor, error) {

	id, err := strconv.ParseUint(filter.Identifier, 10, 64)
	if err != nil {
		return nil, err
	}

	request := Request{
		Action: GetAction,
		Type:   SupervisorRecord,
		Identifiers: RequestIdentifiers{
			Module:     filter.Module,
			Function:   filter.Cluster,
			Supervisor: id,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- request

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (data).(Response)

	if !response.Success {
		return nil, response.Error
	}

	return (response.Data).([]*supervisor.Supervisor), nil
}

func UpdateSupervisor(mandatory Mandatory, data *supervisor.Supervisor) error {

	request := Request{
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

	response := (rsp).(Response)
	return response.Error
}

func FindStatistics(mandatory Mandatory, moduleName, clusterName string) (entries []statistic.Statistics, found bool) {

	databaseRequest := Request{
		Action: GetAction,
		Type:   StatisticRecord,
		Identifiers: RequestIdentifiers{
			Module:   moduleName,
			Function: clusterName,
		},
		Nonce: rand.Uint32(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(Response)

	if !databaseResponse.Success {
		return nil, false
	}

	return databaseResponse.Data.([]statistic.Statistics), true
}

func ShutdownCore(pipe chan<- InterruptEvent) error {
	pipe <- Shutdown
	return nil
}

func GetModules(mandatory Mandatory) (success bool, modules []processor.ModuleData) {

	request := Request{
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

	provisionerResponse := (data).(Response)

	if !provisionerResponse.Success {
		return false, nil
	}

	return true, (provisionerResponse.Data).([]processor.ModuleData)
}

func AddModule(mandatory Mandatory, processorName string, cfg *processor.ModuleConfig) (bool, error) {

	request := Request{
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

	response := (data).(Response)

	return response.Success, response.Error
}

func MountModule(mandatory Mandatory, moduleName string) (bool, error) {

	request := Request{
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

	response := (data).(Response)

	return response.Success, response.Error
}

func UnmountModule(mandatory Mandatory, moduleName string) (bool, error) {

	request := Request{
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

	response := (data).(Response)

	return response.Success, response.Error
}

func DeleteModule(mandatory Mandatory, host string, port int, moduleName string) (bool, error) {

	request := Request{
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

	response := (data).(Response)

	return response.Success, response.Error
}

func FetchFromCache(mandatory Mandatory, key string) (value any, found bool) {

	request := Request{
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

	response := (rsp).(Response)

	return response.Data, response.Success
}

func StoreInCache(mandatory Mandatory, data any, expiry float64) (identifier string, success bool) {

	request := Request{
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

	response := (rsp).(Response)

	cacheResponseData := (response.Data).(CacheResponseData)

	return cacheResponseData.Identifier, response.Success
}

func SwapInCache(mandatory Mandatory, key string, data any) (success bool) {

	request := Request{
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

	response := (rsp).(Response)
	return response.Success
}

func Log(mandatory Mandatory, log *log.Log) error {

	if log == nil {
		return errors.New("need a valid *supervisor.Log")
	}

	request := Request{
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

	response := (rsp).(Response)
	return response.Error
	//return nil
}

func GetJobs(mandatory Mandatory, filter *database.Filter) ([]job.Job, error) {

	request := Request{
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

	response := (rsp).(Response)

	return (response.Data).([]job.Job), nil
}

func CreateJob(mandatory Mandatory, job *job.Job) error {

	request := Request{
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

	response := (rsp).(Response)
	return response.Error
}

func DeleteJob(mandatory Mandatory, filter *database.Filter) error {

	request := Request{
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

	response := (rsp).(Response)
	return response.Error
}

func JobQueue(mandatory Mandatory) ([]job.Job, error) {

	request := Request{
		Action: GetAction,
		Type:   QueueRecord,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response := (rsp).(Response)
	return (response.Data).([]job.Job), response.Error
}

func GetSubscribers(mandatory Mandatory) ([]string, error) {

	request := Request{
		Action: GetAction,
		Type:   SubscriberRecord,
		Nonce:  rand.Uint32(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := multithreaded.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, multithreaded.NoResponseReceived
	}

	response, ok := (rsp).(Response)
	if !ok {
		return nil, InternalError
	}

	subscribers, ok := (response.Data).([]string)
	if !ok {
		return nil, InternalError
	}

	return subscribers, response.Error
}
