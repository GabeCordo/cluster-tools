package thread

import (
	"errors"
	"sync"

	"github.com/GabeCordo/Flock/internal/shared/nonce"
)

var InternalError = errors.New("there was an internal error in the system")

var BadRequestType = errors.New("the request type does not match what was expected for this chan")

var BadResponseType = errors.New("the response type does not match what was expected for this chan")

var IllegalRequest = errors.New("the source is not permitted to send requests over this channel")

var UnknownRequest = errors.New("the request action is unknown to this thread")

type RequestCaller uint8

const (
	User RequestCaller = iota
	System
)

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
	WipeAction
	CloseAction
	ToggleAction
)

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
	ContactRecord
	EmailRecord
	SubscriptionRecord
)

type RequestIdentifiers struct {
	Processor  uint64
	Namespace  string
	Pipeline   string
	Module     string
	Function   string
	Config     string
	Supervisor uint64
}

type Request struct {
	Action      RequestAction
	Type        RequestType
	Identifiers RequestIdentifiers
	Data        any
	Source      Module
	Caller      RequestCaller
	Nonce       nonce.Nonce
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
	Teardown()
}

func SetupListener(in <-chan Request, out chan<- Response, accepting *bool, wg *sync.WaitGroup, module Module, f func(request *Request, response *Response)) {

	go func() {
		for request := range in {
			if !(*accepting) {
				break
			}
			wg.Add(1)

			response := Response{Source: module, Nonce: request.Nonce, Success: false, Error: nil}
			f(&request, &response)

			if out != nil {
				out <- response
			}
			wg.Done()
		}
	}()
}

func Send(request *Request, to chan Request, from Module) {

	request.Source = from
	to <- *request
}
