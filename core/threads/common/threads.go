package common

type RequestAction uint16

type RequestType uint16

type RequestIdentifiers struct {
	Processor  string
	Module     string
	Cluster    string
	Config     string
	Supervisor uint64
}

type TheadRequest struct {
	Action      RequestAction
	Type        RequestType
	Identifiers RequestIdentifiers
	Data        any
	Nonce       uint32
}

type ThreadResponse struct {
	Success bool
	Error   error
	Data    any
	Nonce   uint32
}

type SupervisorAction uint8

const (
	SupervisorGet SupervisorAction = iota
	SupervisorCreate
	SupervisorUpdate
	SupervisorLog
	SupervisorPing
)

type Caller string

const (
	User   Caller = "user"
	System        = "system"
)

type SupervisorRequest struct {
	Action      SupervisorAction
	Identifiers RequestIdentifiers
	Caller      Caller
	Data        any
	Source      Module
	Nonce       uint32
}

type SupervisorResponse struct {
	Success     bool
	Error       error
	Description string
	Data        any
	Nonce       uint32
}

type ProcessorAction uint8

const (
	ProcessorGet ProcessorAction = iota
	ProcessorAdd
	ProcessorRemove
	ProcessorModuleGet
	ProcessorModuleAdd
	ProcessorModuleRemove
	ProcessorModuleMount
	ProcessorModuleUnmount
	ProcessorClusterGet
	ProcessorClusterMount
	ProcessorClusterUnmount
	ProcessorSupervisorGet
	ProcessorSupervisorCreate
	ProcessorSupervisorUpdate
	ProcessorSupervisorLog
	ProcessorPing
)

type ProcessorRequest struct {
	Action      ProcessorAction
	Identifiers RequestIdentifiers
	Data        any
	Source      Module
	Nonce       uint32
}

type ProcessorResponse struct {
	Success     bool
	Error       error
	Description string
	Supervisor  uint64
	Data        any
	Nonce       uint32
}

type CacheAction uint8

const (
	CacheSaveIn CacheAction = iota
	CacheLoadFrom
	CacheWipe
	CacheLowerPing
)

type CacheRequest struct {
	Action CacheAction

	Identifier string
	Data       any
	ExpiresIn  float64 // duration in minutes

	Source Module
	Nonce  uint32
}

type CacheResponse struct {
	Success    bool
	Error      error
	Identifier string
	Data       any
	Nonce      uint32
}

type DatabaseAction uint8

const (
	DatabaseStore DatabaseAction = iota
	DatabaseFetch
	DatabaseReplace
	DatabaseDelete
	DatabaseUpperPing
	DatabaseLowerPing
)

type DatabaseDataType uint8

const (
	SupervisorStatistic DatabaseDataType = iota
	ClusterConfig
	ClusterModule
)

type DatabaseRequest struct {
	Action DatabaseAction `json:"Action"`

	Type    DatabaseDataType `json:"type"`
	Cluster string           `json:"cluster"` // aka. Cluster Identifier
	Module  string           `json:"module"`  // aka. Module Identifier
	Data    any              `json:"data"`    // *cluster.Response `json:"Data"`

	Source Module `json:"origin"`
	Nonce  uint32 `json:"Nonce"`
}

type DatabaseResponse struct {
	Success bool `json:"Success"`
	Error   error
	Data    any    `json:"statistics"` // []database.Entry or cluster.Config
	Nonce   uint32 `json:"Nonce"`
}

type MessengerAction uint8

const (
	MessengerLog MessengerAction = iota
	MessengerWarning
	MessengerFatal
	MessengerClose
	MessengerUpperPing
	MessengerGetSubscribers
	MessengerGetSmtp
	MessengerToggleSmtp
)

type MessengerRequest struct {
	Action MessengerAction `json:"action"`

	Module     string `json:"module"`
	Cluster    string `json:"cluster"`
	Supervisor uint64
	Message    string   `json:"message"`
	Parameters []string `json:"parameters"`

	Data any `json:"data"`

	Source Module `json:"source"`
	Nonce  uint32 `json:"nonce"`
}

type MessengerResponse struct {
	Success bool   `json:"Success"`
	Error   error  `json:"Error"`
	Data    any    `json:"Data"`
	Nonce   uint32 `json:"Nonce"`
}

type SchedulerAction uint8

const (
	SchedulerGet SchedulerAction = iota
	SchedulerCreate
	SchedulerDelete
)

type SchedulerType uint8

const (
	SchedulerJob SchedulerType = iota
	SchedulerQueue
)

type SchedulerRequest struct {
	Action      SchedulerAction
	Type        SchedulerType
	Identifiers RequestIdentifiers
	Data        any
	Nonce       uint32
}

type SchedulerResponse struct {
	Success bool
	Error   error
	Data    any
	Nonce   uint32
}
