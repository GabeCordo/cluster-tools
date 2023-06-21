package core

import (
	"errors"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type MessengerThread struct {
	Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised

	C3  <-chan threads.MessengerRequest  // Messenger is receiving core form the Database
	C4  chan<- threads.MessengerResponse // Messenger is sending responses to the Database
	C11 <-chan threads.MessengerRequest  // Messenger is receiving requests from the Provisioner

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewMessenger(logger *utils.Logger, channels ...interface{}) (*MessengerThread, error) {
	messenger := new(MessengerThread)
	var ok bool

	messenger.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	messenger.C3, ok = (channels[1]).(chan threads.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 1")
	}
	messenger.C4, ok = (channels[2]).(chan threads.MessengerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 2")
	}
	messenger.C11, ok = (channels[3]).(chan threads.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MesengerRequest' in index 3")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	messenger.logger = logger
	messenger.logger.SetColour(utils.Blue)

	return messenger, nil
}
