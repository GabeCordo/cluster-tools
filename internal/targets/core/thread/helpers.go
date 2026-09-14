package thread

import (
	"errors"
	"strconv"

	"github.com/GabeCordo/DistributedFunctions/internal/flags"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/nonce"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/message/log"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/processor"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/run"
)

type Mandatory struct {
	Pipe          chan<- *Request
	ResponseTable *nonce.ResponseTable
	NoncePool     *nonce.Pool
	Log           logging.Logger
	Timeout       float64
}

func GetPipelineFromDatabase(mandatory Mandatory, namespaceName, pipelineName string) (conf []*ScalingFunctions.PipelineIR, found bool) {

	request := Request{
		Action: GetAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(
		mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		return nil, false
	}

	if !databaseResponse.Success {
		return nil, false
	}

	pp, ok := databaseResponse.Data.([]*ScalingFunctions.PipelineIR)
	if !ok {
		return nil, false
	}

	return pp, true
}

func GetNamespacesFromDatabase(mandatory Mandatory) (namespaces []string, err error) {

	request := Request{
		Action: GetAction,
		Type:   NamespaceRecord,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(
		mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		err = nonce.NoResponseReceived
		return namespaces, err
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		err = nonce.InvalidResponseReceived
		return namespaces, err
	}

	if !databaseResponse.Success {
		err = nonce.InvalidResponseReceived
		return namespaces, err
	}

	namespaces, ok = databaseResponse.Data.([]string)
	if !ok {
		err = nonce.InvalidResponseReceived
		return namespaces, err
	}

	return namespaces, err
}

func GetPipelinesFromDatabase(mandatory Mandatory, namespaceName string) (configs []*ScalingFunctions.PipelineIR, found bool) {

	request := Request{
		Action: GetAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(
		mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		return nil, false
	}

	if !databaseResponse.Success {
		return nil, false
	}

	pp, ok := databaseResponse.Data.([]*ScalingFunctions.PipelineIR)
	return pp, ok
}

func StorePipelineInDatabase(mandatory Mandatory, namespace, identifier string, p *ScalingFunctions.PipelineIR) error {

	request := Request{
		Action: CreateAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespace,
			Pipeline:  identifier,
		},
		Data:  p,
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(
		mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		return errors.New("could not cast to *Response")
	}

	// TODO : make the database generate the errors
	if !databaseResponse.Success {
		return errors.New("could not database pipeline in database")
	}

	return nil
}

func ReplacePipelineInDatabase(mandatory Mandatory, namespace, identifier string, p *ScalingFunctions.PipelineIR) (success bool) {

	request := Request{
		Action: UpdateAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespace,
			Pipeline:  identifier,
		},
		Data:  p,
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		success = false
		return success
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		success = false
		return success
	}

	success = databaseResponse.Success
	return success
}

func DeletePipelineInDatabase(mandatory Mandatory, namespaceName, pipelineName string) (success bool) {

	request := Request{
		Action: DeleteAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		success = false
		return success
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		success = false
		return success
	}

	success = databaseResponse.Success
	return success
}

func GetProcessors(mandatory Mandatory) ([]*processor.Processor, bool) {

	request := Request{
		Action: GetAction,
		Type:   ProcessorRecord,
		Source: HttpClient,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	response, ok := (data).(*Response)
	if !ok {
		return nil, false
	}

	if !response.Success {
		return nil, false
	}

	processors, ok := (response.Data).([]*processor.Processor)
	if !ok {
		return nil, false
	}

	return processors, true
}

func AddProcessor(mandatory Mandatory, cfg *processor.Config) (bool, error) {

	request := Request{
		Action: CreateAction,
		Type:   ProcessorRecord,
		Source: Socket,
		Data:   *cfg,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response, ok := (data).(*Response)
	if !ok {
		return false, errors.New("could not cast to *Response")
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response, ok := (data).(*Response)
	if !ok {
		return nonce.InvalidResponseReceived
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse, ok := (data).(*Response)
	if !ok {
		return false
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse, ok := (data).(*Response)
	if !ok {
		return false
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	provisionerResponse, ok := (data).(*Response)
	if !ok {
		return nil, false
	}

	if !provisionerResponse.Success {
		return nil, false
	}

	ff, ok := (provisionerResponse.Data).([]processor.FunctionData)
	return ff, ok
}

func CreateRun(mandatory Mandatory, source Module,
	namespaceName, pipelineName string, metadata map[string]string) (uint64, error) {

	request := Request{
		Action:      CreateAction,
		Type:        RunRecord,
		Identifiers: RequestIdentifiers{Namespace: namespaceName, Pipeline: pipelineName},
		Data:        metadata,
		Source:      source,
		Nonce:       mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return 0, nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return 0, nonce.InvalidResponseReceived
	}

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
			Namespace:  filter.Namespace,
			Pipeline:   filter.Pipeline,
			Supervisor: id,
		},
		Metadata: RequestMetadata{
			Maximum: filter.MaximumResults,
			Offset:  filter.OffsetOfResults,
		},
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
	}

	response, ok := (data).(*Response)
	if !ok {
		return nil, nonce.InvalidResponseReceived
	}

	if !response.Success {
		return nil, response.Error
	}

	runs, ok := (response.Data).([]*run.Run)
	if !ok {
		return nil, errors.New("expected to received []*run.Run")
	}

	return runs, nil
}

func GetRunCount(mandatory Mandatory, filter database.Filter) (count uint32, err error) {

	request := Request{
		Action: CountAction,
		Type:   RunRecord,
		Identifiers: RequestIdentifiers{
			Namespace: filter.Namespace,
			Pipeline:  filter.Pipeline,
		},
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		err = nonce.NoResponseReceived
		return count, err
	}

	response, ok := (data).(*Response)
	if !ok {
		err = nonce.InvalidResponseReceived
		return count, err
	}

	if !response.Success {
		err = response.Error
		return count, err
	}

	count, ok = (response.Data).(uint32)
	if !ok {
		err = errors.New("expected to received uint32")
		return count, err
	}

	return count, err
}

func AsyncUpdateRun(mandatory Mandatory, data *run.Run) {

	request := Request{
		Action: UpdateAction,
		Type:   RunRecord,
		Data:   data,
		Source: Socket,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request
}

func StopRun(mandatory Mandatory, id uint64) error {

	request := Request{
		Action:      DeleteAction,
		Type:        RunRecord,
		Identifiers: RequestIdentifiers{Supervisor: id},
		Nonce:       mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return nonce.InvalidResponseReceived
	}

	return response.Error
}

func FindStatistic(mandatory Mandatory, namespaceName, pipelineName string) (entries []*ScalingFunctions.Statistics, found bool) {

	request := Request{
		Action: GetAction,
		Type:   StatisticRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		return nil, false
	}

	if !databaseResponse.Success {
		return nil, false
	}

	statistics, ok := (databaseResponse.Data).([]*ScalingFunctions.Statistics)
	if !ok {
		return nil, false
	}

	return statistics, true
}

func FindStatistics(mandatory Mandatory, namespaceName string) (fields []string, err error) {

	request := Request{
		Action: SummaryAction,
		Type:   StatisticRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request
	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		err = nonce.NoResponseReceived
		return fields, err
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		err = nonce.InvalidResponseReceived
		return fields, err
	}

	if !databaseResponse.Success {
		return fields, err
	}

	fields, ok = (databaseResponse.Data).([]string)
	if !ok {
		err = nonce.InvalidResponseReceived
		return fields, err
	}

	return fields, err
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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, nil
	}

	provisionerResponse, ok := (data).(*Response)
	if !ok {
		return false, nil
	}

	if !provisionerResponse.Success {
		return false, nil
	}

	modules, success = (provisionerResponse.Data).([]processor.ModuleData)
	return success, modules
}

func AsyncAddModule(mandatory Mandatory, processorId uint64, cfg *ScalingFunctions.ModuleIR) {

	request := new(Request)
	if request == nil {
		panic("failed to allocate Request")
	}

	request.Action = CreateAction
	request.Type = ModuleRecord
	request.Source = Socket
	request.Identifiers = RequestIdentifiers{Processor: processorId}
	request.Data = cfg
	request.Nonce = mandatory.NoncePool.Next()

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response, ok := (data).(*Response)
	if !ok {
		return false, nonce.InvalidResponseReceived
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	data, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response, ok := (data).(*Response)
	if !ok {
		return false, nonce.InvalidResponseReceived
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return nil, false
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return nil, false
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		success = false
	} else {

	}

	response, ok := (rsp).(*Response)
	if !ok {
		return "", false
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return false
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return false
	}

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

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	//HOTFIX : too long to response to the log request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return nonce.InvalidResponseReceived
	}

	return response.Error
}

func GetJobs(mandatory Mandatory, filter *database.Filter) (jobs []*job.Job, err error) {

	jobs = nil

	request := Request{
		Action: GetAction,
		Type:   JobRecord,
		Data:   *filter,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		err = nonce.InvalidResponseReceived
		return jobs, err
	}

	jobs, ok = (response.Data).([]*job.Job)
	if !ok {
		err = errors.New("invalid type received")
		return jobs, err
	}

	return jobs, err
}

func CreateJob(mandatory Mandatory, job *job.Job) error {

	request := Request{
		Action: CreateAction,
		Type:   JobRecord,
		Data:   *job,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return nonce.InvalidResponseReceived
	}

	return response.Error
}

func DeleteJob(mandatory Mandatory, filter *database.Filter) error {

	request := Request{
		Action: DeleteAction,
		Type:   JobRecord,
		Data:   *filter,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return nonce.InvalidResponseReceived
	}

	return response.Error
}

func JobQueue(mandatory Mandatory) ([]*job.Job, error) {

	request := Request{
		Action: GetAction,
		Type:   QueueRecord,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return nil, nonce.InvalidResponseReceived
	}

	jobs, ok := (response.Data).([]*job.Job)
	if !ok {
		return nil, nonce.InvalidResponseReceived
	}

	return jobs, response.Error
}

func GetSubscribers(mandatory Mandatory) ([]string, error) {

	request := Request{
		Action: GetAction,
		Type:   SubscriberRecord,
		Nonce:  mandatory.NoncePool.Next(),
	}

	if flags.DEBUG {
		mandatory.Log.Printf("Sending %s", request.ToString())
	}

	mandatory.Pipe <- &request

	rsp, didTimeout := nonce.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce.NoResponseReceived
	}

	response, ok := (rsp).(*Response)
	if !ok {
		return nil, InternalError
	}

	subscribers, ok := (response.Data).([]string)
	if !ok {
		return nil, InternalError
	}

	return subscribers, response.Error
}

func AsyncGetRun(pipe chan<- *Request, l logging.Logger, n nonce.Nonce, namespace, pipeline string, supervisor, maximum, offset uint64) {

	request := new(Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = GetAction
	request.Type = RunRecord
	request.Identifiers = RequestIdentifiers{
		Namespace:  namespace,
		Pipeline:   pipeline,
		Supervisor: supervisor,
	}
	request.Metadata = RequestMetadata{
		Maximum: maximum,
		Offset:  offset,
	}
	request.Source = Processor
	request.Nonce = n

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncCountRuns(pipe chan<- *Request, l logging.Logger, n nonce.Nonce, namespace, pipeline string) {

	request := new(Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = CountAction
	request.Type = RunRecord
	request.Identifiers = RequestIdentifiers{
		Namespace: namespace,
		Pipeline:  pipeline,
	}
	request.Source = Processor
	request.Nonce = n

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncGetPipeline(pipe chan<- *Request, l logging.Logger, n nonce.Nonce, namespace, pipeline string) {

	request := new(Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = GetAction
	request.Type = PipelineRecord
	request.Identifiers = RequestIdentifiers{
		Namespace: namespace,
		Pipeline:  pipeline,
	}
	request.Source = Processor
	request.Nonce = n

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncCreateRun(pipe chan<- *Request, l logging.Logger, n nonce.Nonce, namespace, module, pipeline string, processor uint64, metadata map[string]string, startedBy Module) {

	request := new(Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = CreateAction
	request.Type = RunRecord
	request.Identifiers = RequestIdentifiers{
		Namespace: namespace,
		Module:    module,
		Pipeline:  pipeline,
	} // will contain the module, cluster
	request.Identifiers.Processor = processor
	request.Caller = User
	request.Data = metadata // will contain the metadata map[string]string
	request.Source = Processor
	request.Nonce = n
	request.StartedBy = startedBy

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	// send the request to the scheduler t
	// the scheduler t will:
	//	1. create a log record of the runner
	//	2. set the log record to the initial state
	//  3. send a provision request to the processor endpoint
	pipe <- request
}

func AsyncUpdateRunToRunner(pipe chan<- *Request, l logging.Logger, oldRequest *Request) {

	request := new(Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = UpdateAction
	request.Type = RunRecord
	request.Identifiers = oldRequest.Identifiers
	request.Data = oldRequest.Data
	request.Source = Processor
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncLogToRunner(pipe chan<- *Request, l logging.Logger, oldRequest *Request) {

	request := new(Request)
	if request == nil {
		panic("failed to allocate thread.Request")
	}

	request.Action = LogAction
	request.Type = RunRecord
	request.Identifiers = oldRequest.Identifiers
	request.Data = oldRequest.Data
	request.Source = Processor
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncStopRunToRunner(pipe chan<- *Request, l logging.Logger, oldRequest *Request) {

	request := new(Request)

	request.Action = DeleteAction
	request.Type = RunRecord
	request.Identifiers = oldRequest.Identifiers
	request.Source = Processor
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncGetPipelineFromDatabase(pipe chan<- *Request, l logging.Logger, oldRequest *Request) {

	request := NewRequest(Runner)

	request.Action = GetAction
	request.Type = PipelineRecord
	request.Identifiers = RequestIdentifiers{
		Namespace: oldRequest.Identifiers.Namespace,
		Pipeline:  oldRequest.Identifiers.Pipeline,
	}
	request.Source = Runner
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncSendRunToSocket(pipe chan<- *Request, l logging.Logger, oldRequest *Request, id uint64, cfg *ScalingFunctions.PipelineIR, metadata map[string]string) {

	runRequest := run.Request{
		Id:        id,
		Namespace: oldRequest.Identifiers.Namespace,
		Config:    cfg,
		Metadata:  metadata,
	}

	request := NewRequest(Runner)

	request.Action = CreateAction
	request.Type = RunRecord
	request.Identifiers = RequestIdentifiers{
		Processor:  oldRequest.Identifiers.Processor,
		Namespace:  oldRequest.Identifiers.Namespace,
		Supervisor: id,
	}
	request.Data = runRequest
	request.Source = Runner
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncCreateStatisticRecordInDatabase(pipe chan<- *Request, l logging.Logger, oldRequest *Request, r *run.Run) {

	req := NewRequest(Runner)

	req.Action = CreateAction
	req.Type = StatisticRecord
	req.Identifiers = RequestIdentifiers{
		Namespace:  r.Namespace,
		Pipeline:   r.Pipeline.Identifier,
		Supervisor: r.Id,
	}
	req.Data = r.GetStatistic()
	req.Source = Runner
	req.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", req.ToString())
	}

	pipe <- req
}

func AsyncCloseMessengerForRun(pipe chan<- *Request, l logging.Logger, oldRequest *Request) {

	request := NewRequest(Runner)

	request.Action = CloseAction
	request.Identifiers = RequestIdentifiers{
		Namespace:  oldRequest.Identifiers.Namespace,
		Pipeline:   oldRequest.Identifiers.Pipeline,
		Supervisor: oldRequest.Identifiers.Supervisor,
	}
	request.Source = Runner
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncSendLogToMessenger(pipe chan<- *Request, l logging.Logger, n nonce.Nonce,
	namespace, pipeline string, identifier uint64,
	t RequestType, message string) {

	request := NewRequest(Runner)

	request.Action = LogAction
	request.Type = t
	request.Identifiers = RequestIdentifiers{
		Namespace:  namespace,
		Pipeline:   pipeline,
		Supervisor: identifier,
	}
	request.Data = message
	request.Nonce = n

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncSendStopToMessenger(pipe chan<- *Request, l logging.Logger, oldRequest *Request, supervisor, processor uint64) {

	request := NewRequest(Runner)

	request.Action = DeleteAction
	request.Type = RunRecord
	request.Identifiers = RequestIdentifiers{
		Supervisor: supervisor,
		Processor:  processor,
	}
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}

func AsyncDeleteRun(pipe chan<- *Request, l logging.Logger, oldRequest *Request, supervisor, processor uint64) {

	request := NewRequest(Runner)

	request.Action = DeleteAction
	request.Type = RunRecord
	request.Identifiers = RequestIdentifiers{
		Supervisor: supervisor,
		Processor:  processor,
	}
	request.Nonce = oldRequest.Nonce

	if flags.DEBUG {
		l.Printf("Sending %s", request.ToString())
	}

	pipe <- request
}
