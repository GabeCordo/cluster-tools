package core

import (
	"context"
	"net/http"
	"sync"
)

// Frontend Thread

const (
	RefreshTime    = 1
	DefaultTimeout = 5
)

type HttpThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan<- DatabaseRequest  // Core is sending core to the Database
	C2 <-chan DatabaseResponse // Core is receiving responses from the Database

	C5 chan<- ProvisionerRequest  // Core is sending core to the Database
	C6 <-chan ProvisionerResponse // Core is receiving responses from the Database

	databaseResponses   map[uint32]DatabaseResponse
	supervisorResponses map[uint32]ProvisionerResponse

	server    *http.Server
	mux       *http.ServeMux
	cancelCtx context.CancelFunc

	accepting bool
	counter   uint32
	mutex     sync.Mutex
	wg        sync.WaitGroup
}

func NewHttp(channels ...interface{}) (*HttpThread, bool) {
	core := new(HttpThread)
	var ok bool

	core.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	core.C1, ok = (channels[1]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	core.C2, ok = (channels[2]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}
	core.C5, ok = (channels[3]).(chan ProvisionerRequest)
	if !ok {
		return nil, ok
	}
	core.C6, ok = (channels[4]).(chan ProvisionerResponse)
	if !ok {
		return nil, ok
	}

	core.databaseResponses = make(map[uint32]DatabaseResponse)
	core.supervisorResponses = make(map[uint32]ProvisionerResponse)

	core.server = new(http.Server)

	core.accepting = true
	core.counter = 0

	return core, ok
}
