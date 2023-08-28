package common

type RequestIdentifiers struct {
	Processor  string
	Module     string
	Cluster    string
	Config     string
	Supervisor uint64
}

type SupervisorAction uint8

const (
	SupervisorFetch SupervisorAction = iota
	SupervisorCreate
	SupervisorUpdate
)

type SupervisorRequest struct {
	Action      SupervisorAction
	Identifiers RequestIdentifiers
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
	ProcessorModuleDelete
	ProcessorModuleMount
	ProcessorModuleUnmount
	ProcessorClusterGet
	ProcessorClusterMount
	ProcessorClusterUnmount
	ProcessorSupervisorFetch
	ProcessorSupervisorCreate
	ProcessorSupervisorUpdate
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
	Identifier string
	Nonce      uint32
	Data       any
	Success    bool
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
	Nonce   uint32 `json:"Nonce"`
	Success bool   `json:"Success"`
	Data    any    `json:"statistics"` // []database.Entry or cluster.Config
}

type MessengerAction uint8

const (
	MessengerLog MessengerAction = iota
	MessengerWarning
	MessengerFatal
	MessengerClose
	MessengerUpperPing
)

type MessengerRequest struct {
	Action MessengerAction `json:"action"`

	Module     string   `json:"module"`
	Cluster    string   `json:"cluster"`
	Message    string   `json:"message"`
	Parameters []string `json:"parameters"`

	Source Module `json:"source"`
	Nonce  uint32 `json:"nonce"`
}

type MessengerResponse struct {
	Nonce   uint32 `json:"Nonce"`
	Success bool   `json:"Success"`
}
