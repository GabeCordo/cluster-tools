package common

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/etl-light/module"
	"github.com/GabeCordo/etl-light/processor"
	"github.com/GabeCordo/etl-light/threads"
)

type SupervisorAction uint8

const (
	SupervisorCreate SupervisorAction = iota
	SupervisorError
	SupervisorClose
)

type SupervisorRequest struct {
	Action      SupervisorAction
	Identifiers struct {
		Module     string
		Cluster    string
		Config     string
		Supervisor uint64
	}
	Source threads.Module
	Nonce  uint32
}

type SupervisorResponse struct {
	Success     bool
	Description string
	Supervisor  uint64
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
	ProcessorClusterProvision
)

type ProcessorRequest struct {
	Action      ProcessorAction
	Identifiers struct {
		Module  string
		Cluster string
		Config  string
	}
	Data struct {
		Cluster   cluster.Config
		Module    module.Config
		Processor processor.Config
	}
	Source threads.Module
	Nonce  uint32
}

type ProcessorResponse struct {
	Success     bool
	Description string
	Supervisor  uint64
	Data        any
	Nonce       uint32
}
