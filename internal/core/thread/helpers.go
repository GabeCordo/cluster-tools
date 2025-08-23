package thread

import (
	"errors"
	"github.com/FortifiedCode/plover"
	"strconv"

	"github.com/FortifiedCode/flock/internal/core/component/message/log"
	processor2 "github.com/FortifiedCode/flock/internal/core/component/processor"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/core/database/job"
	"github.com/FortifiedCode/flock/internal/core/database/run"
	"github.com/FortifiedCode/flock/internal/core/database/statistic"
	nonce2 "github.com/FortifiedCode/flock/internal/shared/nonce"
)

type Mandatory struct {
	Pipe          chan<- *Request
	ResponseTable *nonce2.ResponseTable
	NoncePool     *nonce2.Pool
	Timeout       float64
}

func GetPipelineFromDatabase(mandatory Mandatory, namespaceName, pipelineName string) (conf plover.PipelineIR, found bool) {

	databaseRequest := Request{
		Action: GetAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &databaseRequest

	data, didTimeout := nonce2.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return plover.PipelineIR{}, false
	}

	databaseResponse, ok := (data).(*Response)
	if !ok {
		return plover.PipelineIR{}, false
	}

	if !databaseResponse.Success {
		return plover.PipelineIR{}, false
	}
	return databaseResponse.Data.([]plover.PipelineIR)[0], true
}

func GetPipelinesFromDatabase(mandatory Mandatory, namespaceName string) (configs []plover.PipelineIR, found bool) {

	databaseRequest := Request{
		Action: GetAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &databaseRequest

	data, didTimeout := nonce2.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
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
	return databaseResponse.Data.([]plover.PipelineIR), true
}

func StorePipelineInDatabase(mandatory Mandatory, namespaceName string, p plover.PipelineIR) error {

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
	mandatory.Pipe <- &databaseRequest

	data, didTimeout := nonce2.SendAndWait(
		mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
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

func ReplacePipelineInDatabase(mandatory Mandatory, namespaceName string, p plover.PipelineIR) (success bool) {

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
	mandatory.Pipe <- &databaseRequest

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
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

	databaseRequest := Request{
		Action: DeleteAction,
		Type:   PipelineRecord,
		Identifiers: RequestIdentifiers{
			Namespace: namespaceName,
			Pipeline:  pipelineName,
		},
		Nonce: mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &databaseRequest

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
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

func GetProcessors(mandatory Mandatory) ([]*processor2.Processor, bool) {

	request := Request{
		Action: GetAction,
		Type:   ProcessorRecord,
		Source: HttpClient,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
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

	processors, ok := (response.Data).([]*processor2.Processor)
	if !ok {
		return nil, false
	}

	return processors, true
}

func AddProcessor(mandatory Mandatory, cfg *processor2.Config) (bool, error) {

	request := Request{
		Action: CreateAction,
		Type:   ProcessorRecord,
		Source: Socket,
		Data:   *cfg,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response, ok := (data).(*Response)
	if !ok {
		return false, errors.New("could not cast to *Response")
	}

	return response.Success, response.Error
}

func DeleteProcessor(mandatory Mandatory, cfg *processor2.Config) error {

	request := Request{
		Action: DeleteAction,
		Type:   ProcessorRecord,
		Source: Socket,
		Data:   *cfg,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
	}

	response := (data).(*Response)
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
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(*Response)
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
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false
	}

	provisionerResponse := (data).(*Response)
	return provisionerResponse.Success
}

func GetFunctions(mandatory Mandatory, moduleName string) (clusters []processor2.FunctionData, success bool) {

	request := Request{
		Action:      GetAction,
		Type:        FunctionRecord,
		Identifiers: RequestIdentifiers{Module: moduleName, Function: "", Config: ""},
		Source:      HttpClient,
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	provisionerResponse := (data).(*Response)

	if !provisionerResponse.Success {
		return nil, false
	}

	return (provisionerResponse.Data).([]processor2.FunctionData), true
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
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return 0, nonce2.NoResponseReceived
	}

	response := (rsp).(*Response)

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
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce2.NoResponseReceived
	}

	response := (data).(*Response)

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
	mandatory.Pipe <- &request
}

func StopRun(mandatory Mandatory, id uint64) error {

	request := Request{
		Action:      DeleteAction,
		Type:        RunRecord,
		Identifiers: RequestIdentifiers{Supervisor: id},
		Nonce:       mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
	}

	response := (rsp).(*Response)
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
	mandatory.Pipe <- &databaseRequest

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, databaseRequest.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, false
	}

	databaseResponse := (data).(*Response)

	if !databaseResponse.Success {
		return nil, false
	}

	return databaseResponse.Data.([]statistic.Statistics), true
}

func ShutdownCore(pipe chan<- InterruptEvent) error {
	pipe <- Shutdown
	return nil
}

func GetModules(mandatory Mandatory) (success bool, modules []processor2.ModuleData) {

	request := Request{
		Action: GetAction,
		Type:   ModuleRecord,
		Source: HttpClient,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, nil
	}

	provisionerResponse := (data).(*Response)

	if !provisionerResponse.Success {
		return false, nil
	}

	return true, (provisionerResponse.Data).([]processor2.ModuleData)
}

func AsyncAddModule(mandatory Mandatory, processorId uint64, cfg *plover.ModuleIR) {

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
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(*Response)

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
	mandatory.Pipe <- &request

	data, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return false, errors.New("did not receive a response from the processor thread")
	}

	response := (data).(*Response)

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
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return nil, false
	}

	response := (rsp).(*Response)

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
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		success = false
	} else {

	}

	response := (rsp).(*Response)

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
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)

	if didTimeout {
		return false
	}

	response := (rsp).(*Response)
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
	mandatory.Pipe <- &request

	//HOTFIX : too long to response to the log request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
	}

	response := (rsp).(*Response)
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
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce2.NoResponseReceived
	}

	response := (rsp).(*Response)

	return (response.Data).([]job.Job), nil
}

func CreateJob(mandatory Mandatory, job *job.Job) error {

	request := Request{
		Action: CreateAction,
		Type:   JobRecord,
		Data:   *job,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
	}

	response := (rsp).(*Response)
	return response.Error
}

func DeleteJob(mandatory Mandatory, filter *database.Filter) error {

	request := Request{
		Action: DeleteAction,
		Type:   JobRecord,
		Data:   *filter,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nonce2.NoResponseReceived
	}

	response := (rsp).(*Response)
	return response.Error
}

func JobQueue(mandatory Mandatory) ([]job.Job, error) {

	request := Request{
		Action: GetAction,
		Type:   QueueRecord,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce2.NoResponseReceived
	}

	response := (rsp).(*Response)
	return (response.Data).([]job.Job), response.Error
}

func GetSubscribers(mandatory Mandatory) ([]string, error) {

	request := Request{
		Action: GetAction,
		Type:   SubscriberRecord,
		Nonce:  mandatory.NoncePool.Next(),
	}
	mandatory.Pipe <- &request

	rsp, didTimeout := nonce2.SendAndWait(mandatory.ResponseTable, request.Nonce, mandatory.Timeout)
	if didTimeout {
		return nil, nonce2.NoResponseReceived
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

func AsyncGetRun(pipe chan<- *Request, n nonce2.Nonce, namespace, pipeline string, supervisor uint64) {

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
	request.Source = Processor
	request.Nonce = n

	pipe <- request
}

func AsyncGetPipeline(pipe chan<- *Request, n nonce2.Nonce, namespace, pipeline string) {

	databaseRequest := new(Request)
	if databaseRequest == nil {
		panic("failed to allocate thread.Request")
	}

	databaseRequest.Action = GetAction
	databaseRequest.Type = PipelineRecord
	databaseRequest.Identifiers = RequestIdentifiers{
		Namespace: namespace,
		Pipeline:  pipeline,
	}
	databaseRequest.Source = Processor
	databaseRequest.Nonce = n
	pipe <- databaseRequest
}

func AsyncCreateRun(pipe chan<- *Request, n nonce2.Nonce, namespace, module, pipeline string, processor uint64, metadata map[string]string) {

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

	// send the request to the scheduler t
	// the scheduler t will:
	//	1. create a log record of the runner
	//	2. set the log record to the initial state
	//  3. send a provision request to the processor endpoint
	pipe <- request
}

func AsyncUpdateRunToRunner(pipe chan<- *Request, oldRequest *Request) {

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

	pipe <- request
}

func AsyncLogToRunner(pipe chan<- *Request, oldRequest *Request) {

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

	pipe <- request
}

func AsyncStopRunToRunner(pipe chan<- *Request, oldRequest *Request) {

	request := new(Request)

	request.Action = DeleteAction
	request.Type = RunRecord
	request.Identifiers = oldRequest.Identifiers
	request.Source = Processor
	request.Nonce = oldRequest.Nonce

	pipe <- request
}

func AsyncGetPipelineFromDatabase(pipe chan<- *Request, oldRequest *Request) {

	databaseRequest := NewRequest(Runner)

	databaseRequest.Action = GetAction
	databaseRequest.Type = PipelineRecord
	databaseRequest.Identifiers = RequestIdentifiers{
		Namespace: oldRequest.Identifiers.Namespace,
		Pipeline:  oldRequest.Identifiers.Pipeline,
	}
	databaseRequest.Source = Runner
	databaseRequest.Nonce = oldRequest.Nonce

	pipe <- databaseRequest
}

func AsyncSendRunToSocket(pipe chan<- *Request, oldRequest *Request, id uint64, cfg *plover.PipelineIR, metadata map[string]string) {

	runRequest := run.Request{
		Id:        id,
		Namespace: oldRequest.Identifiers.Namespace,
		Config:    cfg,
		Metadata:  metadata,
	}

	socketRequest := NewRequest(Runner)

	socketRequest.Action = CreateAction
	socketRequest.Type = RunRecord
	socketRequest.Identifiers = RequestIdentifiers{
		Processor:  oldRequest.Identifiers.Processor,
		Namespace:  oldRequest.Identifiers.Namespace,
		Supervisor: id,
	}
	socketRequest.Data = runRequest
	socketRequest.Source = Runner
	socketRequest.Nonce = oldRequest.Nonce

	pipe <- socketRequest
}

func AsyncCreateStatisticRecordInDatabase(pipe chan<- *Request, oldRequest *Request, r *run.Run) {

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

	pipe <- req
}

func AsyncCloseMessengerForRun(pipe chan<- *Request, oldRequest *Request) {

	msgrRequest := NewRequest(Runner)

	msgrRequest.Action = CloseAction
	msgrRequest.Identifiers = RequestIdentifiers{
		Namespace:  oldRequest.Identifiers.Namespace,
		Pipeline:   oldRequest.Identifiers.Pipeline,
		Supervisor: oldRequest.Identifiers.Supervisor,
	}
	msgrRequest.Source = Runner
	msgrRequest.Nonce = oldRequest.Nonce

	pipe <- msgrRequest
}

func AsyncSendLogToMessenger(pipe chan<- *Request, n nonce2.Nonce,
	namespace, pipeline string, identifier uint64,
	t RequestType, message string) {

	messengerRequest := NewRequest(Runner)

	messengerRequest.Action = LogAction
	messengerRequest.Type = t
	messengerRequest.Identifiers = RequestIdentifiers{
		Namespace:  namespace,
		Pipeline:   pipeline,
		Supervisor: identifier,
	}
	messengerRequest.Data = message
	messengerRequest.Nonce = n

	pipe <- messengerRequest
}

func AsyncSendStopToMessenger(pipe chan<- *Request, oldRequest *Request, supervisor, processor uint64) {

	socketRequest := NewRequest(Runner)

	socketRequest.Action = DeleteAction
	socketRequest.Type = RunRecord
	socketRequest.Identifiers = RequestIdentifiers{
		Supervisor: supervisor,
		Processor:  processor,
	}
	socketRequest.Nonce = oldRequest.Nonce

	pipe <- socketRequest
}

func AsyncDeleteRun(pipe chan<- *Request, oldRequest *Request, supervisor, processor uint64) {

	socketRequest := NewRequest(Runner)

	socketRequest.Action = DeleteAction
	socketRequest.Type = RunRecord
	socketRequest.Identifiers = RequestIdentifiers{
		Supervisor: supervisor,
		Processor:  processor,
	}
	socketRequest.Nonce = oldRequest.Nonce

	pipe <- socketRequest
}
