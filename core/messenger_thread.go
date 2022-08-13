package core

import (
	"log"
)

func (msg *MessengerThread) Setup() {
	msg.accepting = true
}

func (msg *MessengerThread) Start() {
	var request MessengerRequest

	// as long as a teardown has not been called, continue looping
	for msg.accepting {
		request = <-msg.C3 // request coming from database
		msg.ProcessIncomingRequest(&request)
	}

	msg.wg.Wait()
}

func (msg *MessengerThread) ProcessIncomingRequest(request *MessengerRequest) {
	switch request.Action {
	case Console:
		log.Println(request.Message)
	}
}

func (msg *MessengerThread) Teardown() {
	msg.accepting = false
}
