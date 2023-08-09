package messenger

import (
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/components/messenger"
	"github.com/GabeCordo/etl/framework/core/common"
)

var messengerInstance *messenger.Messenger

func GetMessengerInstance() *messenger.Messenger {
	if messengerInstance == nil {
		cfg := common.GetConfigInstance()

		messengerInstance = messenger.NewMessenger(
			cfg.Messenger.EnableLogging,
			cfg.Messenger.EnableSmtp,
		)

		if cfg.Messenger.EnableLogging {
			messengerInstance.LoggingDirectory(cfg.Messenger.LogFiles.Directory)
		}

		if cfg.Messenger.EnableSmtp {
			messengerInstance.SetupSMTP(
				messenger.Endpoint{
					Host: cfg.Messenger.Smtp.Endpoint.Host,
					Port: cfg.Messenger.Smtp.Endpoint.Port,
				},
				messenger.Credentials{
					Email:    cfg.Messenger.Smtp.Credentials.Email,
					Password: cfg.Messenger.Smtp.Credentials.Password,
				},
			).SetupReceivers(
				cfg.Messenger.Smtp.Subscribers,
			)
		}
	}
	return messengerInstance
}

func (messengerThread *Thread) Setup() {
	messengerThread.accepting = true
}

func (messengerThread *Thread) Start() {
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

func (messengerThread *Thread) Respond(module threads.Module, response any) (success bool) {

	success = true

	switch module {
	case threads.Database:
		messengerThread.C4 <- *(response).(*threads.MessengerResponse)
	default:
		success = false
	}

	return success
}

func (messengerThread *Thread) ProcessIncomingRequest(request *threads.MessengerRequest) {

	switch request.Action {
	case threads.MessengerClose:
		messengerThread.ProcessCloseLogRequest(request)
	case threads.MessengerUpperPing:
		messengerThread.ProcessMessengerPing(request)
	default:
		messengerThread.ProcessConsoleRequest(request)
	}

	messengerThread.wg.Done()
}

func (messengerThread *Thread) ProcessMessengerPing(request *threads.MessengerRequest) {

	if common.GetConfigInstance().Debug {
		messengerThread.logger.Println("[etl_messenger] received ping over C3")
	}

	response := &threads.MessengerResponse{Nonce: request.Nonce, Success: true}
	messengerThread.Respond(threads.Database, response)
}

func (messengerThread *Thread) ProcessConsoleRequest(request *threads.MessengerRequest) {
	messengerInstance := GetMessengerInstance()

	var priority messenger.MessagePriority

	switch request.Action {
	case threads.MessengerLog:
		priority = messenger.Log
	case threads.MessengerWarning:
		priority = messenger.Warning
	default:
		priority = messenger.Fatal
	}

	messengerInstance.Log(request.Cluster, request.Message, priority)
}

func (messengerThread *Thread) ProcessCloseLogRequest(request *threads.MessengerRequest) {
	messengerInstance := GetMessengerInstance()
	messengerInstance.Complete(request.Cluster)
}

func (messengerThread *Thread) Teardown() {
	messengerThread.accepting = false

	messengerThread.wg.Wait()
}
