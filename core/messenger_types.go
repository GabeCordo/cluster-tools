package core

import "sync"

type MessengerAction uint8

const (
	Log MessengerAction = iota
	Warning
	Fatal
	Close
	MessengerPing
)

type MessengerRequest struct {
	Action     MessengerAction `json:"action"`
	Cluster    string          `json:"cluster"`
	Nonce      uint32          `json:"nonce"`
	Message    string          `json:"message"`
	Parameters []string        `json:"parameters"`
}

type MessengerResponse struct {
	Nonce uint32 `json:"Nonce"`
}

type MessengerThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C3  <-chan MessengerRequest  // Messenger is receiving core form the Database
	C4  chan<- MessengerResponse // Messenger is sending responses to the Database
	C11 <-chan MessengerRequest  // Messenger is receiving requests from the Provisioner

	accepting bool
	wg        sync.WaitGroup
}

func NewMessenger(channels ...interface{}) (*MessengerThread, bool) {
	messenger := new(MessengerThread)
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
	messenger.C11, ok = (channels[3]).(chan MessengerRequest)
	if !ok {
		return nil, ok
	}

	return messenger, ok
}
