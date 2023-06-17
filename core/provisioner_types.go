package core

import (
	"errors"
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type ProvisionerAction int8

const (
	ProvisionerProvision ProvisionerAction = iota
	ProvisionerModuleLoad
	ProvisionerModuleDelete
	ProvisionerDynamicLoad
	ProvisionerDynamicDelete
	ProvisionerMount
	ProvisionerUnMount
	ProvisionerTeardown
	ProvisionerLowerPing
)

type ProvisionerRequest struct {
	Action      ProvisionerAction `json:"Action"`
	Nonce       uint32            `json:"Nonce"`
	ModuleName  string            `json:"Module"`
	ClusterName string            `json:"cluster"`
	Mount       bool              `json:"mount,omitempty"`
	Config      string            `json:"config,omitempty"`
	ModulePath  string            `json:"path,omitempty"`
}

type ProvisionerResponse struct {
	Nonce        uint32 `json:"nonce"`
	Success      bool   `json:"success"`
	Cluster      string `json:"cluster"`
	Description  string `json:"description"`
	SupervisorId uint64 `json:"supervisor-id"`
}

type ProvisionerThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C5 <-chan ProvisionerRequest  // Supervisor is receiving core from the http_thread
	C6 chan<- ProvisionerResponse // Supervisor is sending responses to the http_thread

	C7 chan<- DatabaseRequest  // Supervisor is sending core to the database
	C8 <-chan DatabaseResponse // Supervisor is receiving responses from the database

	C9  chan<- CacheRequest  // Provisioner is sending requests to the cache
	C10 <-chan CacheResponse // Provisioner is receiving responses from the CacheThread

	C11 chan<- MessengerRequest // Provisioner is sending request to the messenger

	databaseResponseTable *utils.ResponseTable
	cacheResponseTable    *utils.ResponseTable

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewProvisioner(logger *utils.Logger, channels ...interface{}) (*ProvisionerThread, error) {
	provisioner := new(ProvisionerThread)
	var ok bool

	provisioner.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	provisioner.C5, ok = (channels[1]).(chan ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	provisioner.C6, ok = (channels[2]).(chan ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}
	provisioner.C7, ok = (channels[3]).(chan DatabaseRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	provisioner.C8, ok = (channels[4]).(chan DatabaseResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	provisioner.C9, ok = (channels[5]).(chan CacheRequest)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 5")
	}
	provisioner.C10, ok = (channels[6]).(chan CacheResponse)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 6")
	}
	provisioner.C11, ok = (channels[7]).(chan MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}

	provisioner.databaseResponseTable = utils.NewResponseTable()
	provisioner.cacheResponseTable = utils.NewResponseTable()

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	provisioner.logger = logger
	provisioner.logger.SetColour(utils.Orange)

	return provisioner, nil
}
