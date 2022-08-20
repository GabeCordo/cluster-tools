package core

import "sync"

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

	C9  chan<- StateMachineRequest  // Core is sending requests to the state machine
	C10 <-chan StateMachineResponse // Core is receiving responses from the state machine

	databaseResponses   map[uint32]DatabaseResponse
	supervisorResponses map[uint32]ProvisionerResponse

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
	core.C9, ok = (channels[5]).(chan StateMachineRequest)
	if !ok {
		return nil, ok
	}
	core.C10, ok = (channels[6]).(chan StateMachineResponse)
	if !ok {
		return nil, ok
	}

	core.databaseResponses = make(map[uint32]DatabaseResponse)
	core.supervisorResponses = make(map[uint32]ProvisionerResponse)
	core.accepting = true
	core.counter = 0

	return core, ok
}
