package thread

import (
	"errors"
	"fmt"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/nonce"
)

var InternalError = errors.New("there was an internal error in the system")

var BadRequestType = errors.New("the request type does not match what was expected for this chan")

var BadResponseType = errors.New("the response type does not match what was expected for this chan")

var UnknownRequest = errors.New("the request action is unknown to this thread")

var NotImplemented = errors.New("thread functions has not been implemented")

var FailedToSendResponse = errors.New("could not send a response")

type RequestCaller uint8

const (
	User RequestCaller = iota
)

const NumOfRequestActions = 11

type RequestAction uint16

const (
	PingAction RequestAction = iota
	GetAction
	CreateAction
	UpdateAction
	DeleteAction
	LogAction
	MountAction
	UnMountAction
	CloseAction
	SummaryAction
	CountAction
)

var requestActionStrings = [NumOfRequestActions]string{
	"PingAction",
	"GetAction",
	"CreateAction",
	"UpdateAction",
	"DeleteAction",
	"LogAction",
	"MountAction",
	"UnMountAction",
	"CloseAction",
	"SummaryAction",
	"CountAction",
}

const NumOfRequestTypes = 15

type RequestType uint16

const (
	ProcessorRecord RequestType = iota
	ModuleRecord
	FunctionRecord
	RunRecord
	PipelineRecord
	CacheRecord
	SmtpRecord
	JobRecord
	QueueRecord
	StatisticRecord
	DefaultLogRecord
	WarningLogRecord
	FatalLogRecord
	SubscriberRecord
	NamespaceRecord
)

var requestTypeStrings = [NumOfRequestTypes]string{
	"ProcessorRecord",
	"ModuleRecord",
	"FunctionRecord",
	"RunRecord",
	"PipelineRecord",
	"CacheRecord",
	"SmtpRecord",
	"JobRecord",
	"QueueRecord",
	"StatisticRecord",
	"DefaultLogRecord",
	"WarningLogRecord",
	"FatalLogRecord",
	"SubscriberRecord",
	"NamespaceRecord",
}

type RequestIdentifiers struct {
	Processor  uint64
	Namespace  string
	Pipeline   string
	Module     string
	Function   string
	Config     string
	Supervisor uint64
}

type RequestMetadata struct {
	Maximum uint64
	Offset  uint64
}

type Request struct {
	Action      RequestAction
	Type        RequestType
	Identifiers RequestIdentifiers
	Metadata    RequestMetadata
	Data        any
	Source      Module
	StartedBy   Module
	Caller      RequestCaller
	Nonce       nonce.Nonce
}

func (request Request) ToString() string {
	return fmt.Sprintf(
		"Request(action: %s, type: %s, nonce: %d)",
		requestActionStrings[request.Action],
		requestTypeStrings[request.Type],
		request.Nonce,
	)
}

type Response struct {
	Action  RequestAction
	Type    RequestType
	Success bool
	Error   error
	Data    any
	Source  Module
	Nonce   nonce.Nonce
}

func (response Response) ToString() string {
	return fmt.Sprintf(
		"Response(success: %t, err_nil: %t, action: %s, type: %s, nonce: %d)",
		response.Success,
		response.Error == nil,
		requestActionStrings[response.Action],
		requestTypeStrings[response.Type],
		response.Nonce,
	)
}

type ProcessorResponseData struct {
	Supervisor uint64
	Data       any
}

type CacheRequestData struct {
	Data       any
	Identifier string
	ExpiresIn  float64 // duration in minutes
}

type CacheResponseData struct {
	Identifier string
	Data       any
}

type DatabaseRequestData struct {
	Cluster string `json:"cluster"` // aka. Cluster Identifier
	Module  string `json:"module"`  // aka. Module Identifier
	Data    any    `json:"data"`    // *cluster.Response `json:"Data"`
}

type MessengerRequestData struct {
	Module     string `json:"module"`
	Cluster    string `json:"cluster"`
	Supervisor uint64
	Message    string   `json:"message"`
	Parameters []string `json:"parameters"`
	Data       any      `json:"data"`
}

type InterruptEvent uint8

const (
	Shutdown InterruptEvent = 0
	Panic                   = 1
)

type Module uint8

const (
	HttpClient Module = iota
	Socket
	Database
	Processor
	Runner
	Messenger
	Cache
	Scheduler
)

type Thread interface {
	Setup()
	Start()
	HandleRequest(*Request) *Response
	TearDown()
}

func NewRequest(source Module) *Request {
	request := new(Request)
	if request == nil {
		panic("failed to allocated thread.Request struct")
	}
	request.Source = source
	return request
}

func NewResponse(source Module) *Response {
	response := new(Response)
	if response == nil {
		panic("failed to allocated thread.Response struct")
	}
	response.Source = source
	return response
}

func CopyMetadata(request *Request, response *Response) {

	response.Action = request.Action
	response.Type = request.Type
	response.Nonce = request.Nonce
}
