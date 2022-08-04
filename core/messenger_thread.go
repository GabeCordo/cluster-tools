package core

import (
	"log"
)

func (msg *Messenger) Setup() {
	msg.accepting = true
}

func (msg *Messenger) Start() {
	var request MessengerRequest

	// as long as a teardown has not been called, continue looping
	for msg.accepting {
		request = <-msg.C3 // request coming from database

		switch request.action {
		case Console:
			log.Println(request.message)
		}
	}

	msg.wg.Wait()
}

func (msg *Messenger) Teardown() {
	msg.accepting = false
}
