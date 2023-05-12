package core

import (
	"github.com/GabeCordo/etl/components/messenger"
)

var messengerInstance *messenger.Messenger

func GetMessengerInstance() *messenger.Messenger {
	if messengerInstance == nil {
		config := GetConfigInstance()

		messengerInstance = messenger.NewMessenger(
			config.Messenger.EnableLogging,
			config.Messenger.EnableSmtp,
		)

		if config.Messenger.EnableLogging {
			messengerInstance.LoggingDirectory(config.Messenger.LogFiles.Directory)
		}

		if config.Messenger.EnableSmtp {
			messengerInstance.SetupSMTP(
				config.Messenger.Smtp.Endpoint,
				config.Messenger.Smtp.Credentials,
			).SetupReceivers(
				config.Messenger.Smtp.Subscribers,
			)
		}
	}
	return messengerInstance
}

func (msg *MessengerThread) Setup() {
	msg.accepting = true
}

func (msg *MessengerThread) Start() {
	// as long as a teardown has not been called, continue looping

	go func() {
		// request coming from database
		for request := range msg.C3 {
			if !msg.accepting {
				break
			}
			msg.wg.Add(1)
			msg.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// request coming from provisioner
		for request := range msg.C11 {
			if !msg.accepting {
				break
			}
			msg.wg.Add(1)
			msg.ProcessIncomingRequest(&request)
		}
	}()

	msg.wg.Wait()
}

func (msg *MessengerThread) ProcessIncomingRequest(request *MessengerRequest) {
	if request.Action == Close {
		msg.ProcessCloseLogRequest(request)
	} else if request.Action == MessengerPing {
		msg.ProcessMessengerPing(request)
	} else {
		msg.ProcessConsoleRequest(request)
	}
	msg.wg.Done()
}

func (msg *MessengerThread) ProcessMessengerPing(request *MessengerRequest) {

}

func (msg *MessengerThread) ProcessConsoleRequest(request *MessengerRequest) {
	messengerInstance := GetMessengerInstance()

	var priority messenger.MessagePriority
	if request.Action == Log {
		priority = messenger.Log
	} else if request.Action == Warning {
		priority = messenger.Warning
	} else {
		priority = messenger.Fatal
	}

	messengerInstance.Log(request.Cluster, request.Message, priority)
}

func (msg *MessengerThread) ProcessCloseLogRequest(request *MessengerRequest) {
	messenger := GetMessengerInstance()
	messenger.Complete(request.Cluster)
}

func (msg *MessengerThread) Teardown() {
	msg.accepting = false

	msg.wg.Wait()
}
