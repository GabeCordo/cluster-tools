package core

import "sync"

type StateMachineRequest struct {
}

type StateMachineResponse struct {
}

type StateMachineThread struct {
	C9  <-chan StateMachineRequest
	C10 chan<- StateMachineResponse

	C11 <-chan StateMachineRequest
	C12 chan<- StateMachineResponse

	accepting bool
	wg        sync.WaitGroup
	mutex     sync.Mutex
}
