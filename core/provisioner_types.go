package core

import "sync"

type SupervisorAction int8

const (
	Provision SupervisorAction = iota
	Mount
	UnMount
	Teardown
)

type ProvisionerRequest struct {
	Action     SupervisorAction `json:"action"`
	Nonce      uint32           `json:"nonce"`
	Cluster    string           `json:"cluster"`
	Parameters []string         `json"parameters"`
}

type ProvisionerResponse struct {
	Nonce   uint32 `json:"nonce"`
	Success bool   `json:"success"`
}

type ProvisionerThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C5 <-chan ProvisionerRequest  // Supervisor is receiving core from the http_thread
	C6 chan<- ProvisionerResponse // Supervisor is sending responses to the http_thread

	C7 chan<- DatabaseRequest  // Supervisor is sending core to the database
	C8 <-chan DatabaseResponse // Supervisor is receiving responses from the database

	accepting bool
	wg        sync.WaitGroup
}

func NewProvisioner(channels ...interface{}) (*ProvisionerThread, bool) {
	provisioner := new(ProvisionerThread)
	var ok bool

	provisioner.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	provisioner.C5, ok = (channels[1]).(chan ProvisionerRequest)
	if !ok {
		return nil, ok
	}
	provisioner.C6, ok = (channels[2]).(chan ProvisionerResponse)
	if !ok {
		return nil, ok
	}
	provisioner.C7, ok = (channels[3]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	provisioner.C8, ok = (channels[4]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}

	return provisioner, ok
}
