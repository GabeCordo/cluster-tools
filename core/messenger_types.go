package core

import (
	"errors"
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type MessengerAction uint8

const (
	MessengerLog MessengerAction = iota
	MessengerWarning
	MessengerFatal
	MessengerClose
	MessengerUpperPing
)

type MessengerRequest struct {
	Action     MessengerAction `json:"action"`
	Cluster    string          `json:"cluster"`
	Nonce      uint32          `json:"nonce"`
	Message    string          `json:"message"`
	Parameters []string        `json:"parameters"`
}

type MessengerResponse struct {
	Nonce   uint32 `json:"Nonce"`
	Success bool   `json:"Success"`
}

type MessengerThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C3  <-chan MessengerRequest  // Messenger is receiving core form the Database
	C4  chan<- MessengerResponse // Messenger is sending responses to the Database
	C11 <-chan MessengerRequest  // Messenger is receiving requests from the Provisioner

	logger *utils.Logger

	accepting bool
	wg        sync.WaitGroup
}

func NewMessenger(logger *utils.Logger, channels ...interface{}) (*MessengerThread, error) {
	messenger := new(MessengerThread)
	var ok bool

	messenger.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	messenger.C3, ok = (channels[1]).(chan MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 1")
	}
	messenger.C4, ok = (channels[2]).(chan MessengerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 2")
	}
	messenger.C11, ok = (channels[3]).(chan MessengerRequest)
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
