package common

import (
	"github.com/GabeCordo/mango/threads"
)

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
	Source      threads.Module
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
	Source      threads.Module
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
