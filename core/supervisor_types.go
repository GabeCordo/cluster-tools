package core

import "sync"

type SupervisorAction int8

const (
	Provision SupervisorAction = 0
	Teardown                   = 1
)

type SupervisorRequest struct {
	Action     SupervisorAction `json:"action"`
	Nonce      uint32           `json:"nonce"`
	Cluster    string           `json:"cluster"`
	Parameters []string         `json"parameters"`
}

type SupervisorResponse struct {
	Nonce   uint32 `json:"nonce"`
	Success bool   `json:"success"`
}

type SupervisorThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C5 <-chan SupervisorRequest  // Supervisor is receiving core from the http_thread
	C6 chan<- SupervisorResponse // Supervisor is sending responses to the http_thread

	C7 chan<- DatabaseRequest  // Supervisor is sending core to the database
	C8 <-chan DatabaseResponse // Supervisor is receiving responses from the database

	accepting bool
	wg        sync.WaitGroup
}

func NewSupervisor(channels ...interface{}) (*SupervisorThread, bool) {
	supervisor := new(SupervisorThread)
	var ok bool

	supervisor.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	supervisor.C5, ok = (channels[1]).(chan SupervisorRequest)
	if !ok {
		return nil, ok
	}
	supervisor.C6, ok = (channels[2]).(chan SupervisorResponse)
	if !ok {
		return nil, ok
	}
	supervisor.C7, ok = (channels[3]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	supervisor.C8, ok = (channels[4]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}

	return supervisor, ok
}
