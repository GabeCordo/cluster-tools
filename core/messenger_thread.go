package core

import (
	"log"
)

func (msg *MessengerThread) Setup() {
	msg.accepting = true
}

func (msg *MessengerThread) Start() {
	// as long as a teardown has not been called, continue looping

	// request coming from database
	for request := range msg.C3 {
		if !msg.accepting {
			break
		}
		msg.wg.Add(1)
		msg.ProcessIncomingRequest(&request)
	}

	msg.wg.Wait()
}

func (msg *MessengerThread) ProcessIncomingRequest(request *MessengerRequest) {
	switch request.Action {
	case Console:
		log.Println(request.Message)
	}
	msg.wg.Done()
}

func (msg *MessengerThread) Teardown() {
	msg.accepting = false

	msg.wg.Wait()
}
