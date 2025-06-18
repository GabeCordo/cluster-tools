package thread

import (
	"errors"
	"strconv"

	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/job"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/database/statistic"
	"github.com/GabeCordo/Flock/internal/core/message/log"
	"github.com/GabeCordo/Flock/internal/core/processor"
	"github.com/GabeCordo/Flock/internal/nonce"
	"github.com/GabeCordo/toolchain/multithreaded"
)

type Mandatory struct {
	Pipe          chan<- Request
	ResponseTable *nonce.ResponseTable
	NoncePool     *nonce.Pool
	Timeout       float64
}

func GetPipelineFromDatabase(mandatory Mandatory, namespaceName, pipelineName string) (conf pipeline.Pipeline, found bool) {

	databaseRequest := Request{
		Action: GetAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := nonce.SendAndWait(
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

func GetPipelinesFromDatabase(mandatory Mandatory, namespaceName string) (configs []pipeline.Pipeline, found bool) {

	databaseRequest := Request{
		Action: GetAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := nonce.SendAndWait(
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

func StorePipelineInDatabase(mandatory Mandatory, namespaceName string, p pipeline.Pipeline) error {

	databaseRequest := Request{
		Action: CreateAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  p.Identifier,
		},
		Data:  p,
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := nonce.SendAndWait(
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

func ReplacePipelineInDatabase(mandatory Mandatory, namespaceName string, p pipeline.Pipeline) (success bool) {

	databaseRequest := Request{
		Action: UpdateAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  p.Identifier,
		},
		Data:  p,
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	databaseResponse := (data).(Response)
	return databaseResponse.Success
}

func DeletePipelineInDatabase(mandatory Mandatory, namespaceName, pipelineName string) (success bool) {

	databaseRequest := Request{
		Action: DeleteAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
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
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
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
		Source: Socket,
		Data:   *cfg,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
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
		Source: Socket,
		Data:   *cfg,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response := (data).(Response)
	return response.Error
}

func MountFunction(mandatory Mandatory, moduleName, functionName string) (success bool) {

	request := Request{
		Action:      MountAction,
		Type:        FunctionRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Function: functionName, Config: ""},
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(Response)
	return provisionerResponse.Success
}

func UnmountFunction(mandatory Mandatory, moduleName, functionName string) (success bool) {

	request := Request{
		Action:      UnMountAction,
		Type:        FunctionRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName, Function: functionName, Config: ""},
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
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
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(Response)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).([]processor.FunctionData), true
}

func CreateRun(mandatory Mandatory,
	namespaceName, pipelineName string, metadata map[string]string) (uint64, error) {

	request := Request{
		Action:      CreateAction,
		Type:        RunRecord,
		Identifiers: RequestIdentifiers{Namespace: namespaceName, Pipeline: pipelineName},
		Data:        metadata,
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return 0, nonce.NoResponseReceived
	}

	response := (rsp).(Response)

	id, _ := response.Data.(uint64)
	return id, response.Error
}

func GetRun(mandatory Mandatory, filter database.Filter) ([]*run.Run, error) {

	id, err := strconv.ParseUint(filter.Identifier, 10, 64)
	if err != nil {
		return nil, err
	}

	request := Request{
		Action: GetAction,
		Type:   RunRecord,
		Identifiers: RequestIdentifiers{
			Module:     filter.Namespace,
			Function:   filter.Pipeline,
			Supervisor: id,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
	}

	response := (data).(Response)

	if !response.Success {
		return nil, response.Error
	}

	return (response.Data).([]*run.Run), nil
}

func AsyncUpdateRun(mandatory Mandatory, data *run.Run) {

	request := Request{
		Action: UpdateAction,
		Type:   RunRecord,
		Data:   data,
		Source: Socket,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request
}

func StopRun(mandatory Mandatory, id uint64) error {

	request := Request{
		Action:      DeleteAction,
		Type:        RunRecord,
		Identifiers: RequestIdentifiers{Supervisor: id},
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response := (rsp).(Response)
	return response.Error
}

func FindStatistics(mandatory Mandatory, namespaceName, pipelineName string) (entries []statistic.Statistics, found bool) {

	databaseRequest := Request{
		Action: GetAction,
		Type:   StatisticRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- databaseRequest

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
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
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, nil
	}

	provisionerResponse := (data).(Response)

	if !provisionerResponse.Success {
		return false, nil
	}

	return true, (provisionerResponse.Data).([]processor.ModuleData)
}

func AsyncAddModule(mandatory Mandatory, processorId uint64, cfg *processor.ModuleConfig) {

	request := Request{
		Action:      CreateAction,
		Type:        ModuleRecord,
		Source:      Socket,
		Identifiers: RequestIdentifiers{Processor: processorId},
		Data:        *cfg,
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request
}

func MountModule(mandatory Mandatory, moduleName string) (bool, error) {

	request := Request{
		Action:      MountAction,
		Type:        ModuleRecord,
		Source:      HttpClient,
		Identifiers: RequestIdentifiers{Module: moduleName},
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
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
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(Response)

	return response.Success, response.Error
}

func DeleteModule(mandatory Mandatory, processorId uint64, moduleName string) (bool, error) {

	request := Request{
		Action:      DeleteAction,
		Type:        ModuleRecord,
		Source:      Socket,
		Identifiers: RequestIdentifiers{Processor: processorId, Module: moduleName},
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return false, nonce.NoResponseReceived
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
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

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
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

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
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return false
	}

	response := (rsp).(Response)
	return response.Success
}

func Log(mandatory Mandatory, log *log.Log) error {

	if log == nil {
		return errors.New("need a valid *runner.Log")
	}

	request := Request{
		Action: LogAction,
		Type:   RunRecord,
		Data:   log,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	//HOTFIX : too long to response to the log request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
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
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
	}

	response := (rsp).(Response)

	return (response.Data).([]job.Job), nil
}

func CreateJob(mandatory Mandatory, job *job.Job) error {

	request := Request{
		Action: CreateAction,
		Type:   JobRecord,
		Data:   *job,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response := (rsp).(Response)
	return response.Error
}

func DeleteJob(mandatory Mandatory, filter *database.Filter) error {

	request := Request{
		Action: DeleteAction,
		Type:   JobRecord,
		Data:   *filter,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response := (rsp).(Response)
	return response.Error
}

func JobQueue(mandatory Mandatory) ([]job.Job, error) {

	request := Request{
		Action: GetAction,
		Type:   QueueRecord,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
	}

	response := (rsp).(Response)
	return (response.Data).([]job.Job), response.Error
}

func GetSubscribers(mandatory Mandatory) ([]string, error) {

	request := Request{
		Action: GetAction,
		Type:   SubscriberRecord,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
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
