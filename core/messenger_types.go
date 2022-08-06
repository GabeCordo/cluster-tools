package core

import "sync"

type MessengerAction uint8

const (
	Console MessengerAction = 0
)

type MessengerRequest struct {
	action     MessengerAction
	message    string
	parameters []string
}

type MessengerResponse struct {
}

type Messenger struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C3 <-chan MessengerRequest  // Messenger is receiving core form the Database
	C4 chan<- MessengerResponse // Messenger is sending responses to the Database

	accepting bool
	wg        sync.WaitGroup
}

func NewMessenger(channels ...interface{}) (*Messenger, bool) {
	messenger := new(Messenger)
	var ok bool

	messenger.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	messenger.C3, ok = (channels[1]).(chan MessengerRequest)
	if !ok {
		return nil, ok
	}
	messenger.C4, ok = (channels[2]).(chan MessengerResponse)
	if !ok {
		return nil, ok
	}

	return messenger, ok
}
