package core

import (
	"github.com/GabeCordo/etl-light/core/threads"
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
				messenger.Endpoint{
					Host: config.Messenger.Smtp.Endpoint.Host,
					Port: config.Messenger.Smtp.Endpoint.Port,
				},
				messenger.Credentials{
					Email:    config.Messenger.Smtp.Credentials.Email,
					Password: config.Messenger.Smtp.Credentials.Password,
				},
			).SetupReceivers(
				config.Messenger.Smtp.Subscribers,
			)
		}
	}
	return messengerInstance
}

func (messengerThread *MessengerThread) Setup() {
	messengerThread.accepting = true
}

func (messengerThread *MessengerThread) Start() {
	// as long as a teardown has not been called, continue looping

	go func() {
		// request coming from database
		for request := range messengerThread.C3 {
			if !messengerThread.accepting {
				break
			}
			messengerThread.wg.Add(1)
			messengerThread.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// request coming from provisioner
		for request := range messengerThread.C11 {
			if !messengerThread.accepting {
				break
			}
			messengerThread.wg.Add(1)
			messengerThread.ProcessIncomingRequest(&request)
		}
	}()

	messengerThread.wg.Wait()
}

func (messengerThread *MessengerThread) Send(module threads.Module, response *threads.MessengerResponse) {

	messengerThread.C4 <- *response
}

func (messengerThread *MessengerThread) ProcessIncomingRequest(request *threads.MessengerRequest) {

	if request.Action == threads.MessengerClose {
		messengerThread.ProcessCloseLogRequest(request)
	} else if request.Action == threads.MessengerUpperPing {
		messengerThread.ProcessMessengerPing(request)
	} else {
		messengerThread.ProcessConsoleRequest(request)
	}

	messengerThread.wg.Done()
}

func (messengerThread *MessengerThread) ProcessMessengerPing(request *threads.MessengerRequest) {

	if GetConfigInstance().Debug {
		messengerThread.logger.Println("[etl_messenger] received ping over C3")
	}

	messengerThread.C4 <- threads.MessengerResponse{Nonce: request.Nonce, Success: true}
}

func (messengerThread *MessengerThread) ProcessConsoleRequest(request *threads.MessengerRequest) {
	messengerInstance := GetMessengerInstance()

	var priority messenger.MessagePriority
	if request.Action == threads.MessengerLog {
		priority = messenger.Log
	} else if request.Action == threads.MessengerWarning {
		priority = messenger.Warning
	} else {
		priority = messenger.Fatal
	}

	messengerInstance.Log(request.Cluster, request.Message, priority)
}

func (messengerThread *MessengerThread) ProcessCloseLogRequest(request *threads.MessengerRequest) {
	messenger := GetMessengerInstance()
	messenger.Complete(request.Cluster)
}

func (messengerThread *MessengerThread) Teardown() {
	messengerThread.accepting = false

	messengerThread.wg.Wait()
}
