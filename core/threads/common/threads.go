package common

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
	ClusterRecord
	SupervisorRecord
	ConfigRecord
	CacheRecord
	SmtpRecord
	JobRecord
	QueueRecord
	StatisticRecord
	DefaultLogRecord
	WarningLogRecord
	FatalLogRecord
	SubscriberRecord
)

type RequestIdentifiers struct {
	Processor  string
	Module     string
	Cluster    string
	Config     string
	Supervisor uint64
}

type ThreadRequest struct {
	Action      RequestAction
	Type        RequestType
	Identifiers RequestIdentifiers
	Data        any
	Source      Module
	Caller      RequestCaller
	Nonce       uint32
}

type ThreadResponse struct {
	Success bool
	Error   error
	Data    any
	Source  Module
	Nonce   uint32
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
