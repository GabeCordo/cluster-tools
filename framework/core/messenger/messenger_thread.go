package messenger

import (
	"fmt"
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

func (messengerThread *Thread) Send(module threads.Module, response *threads.MessengerResponse) {

	messengerThread.C4 <- *response
}

func (messengerThread *Thread) ProcessIncomingRequest(request *threads.MessengerRequest) {

	if request.Action == threads.MessengerClose {
		messengerThread.ProcessCloseLogRequest(request)
	} else if request.Action == threads.MessengerUpperPing {
		messengerThread.ProcessMessengerPing(request)
	} else {
		messengerThread.ProcessConsoleRequest(request)
	}

	messengerThread.wg.Done()
}

func (messengerThread *Thread) ProcessMessengerPing(request *threads.MessengerRequest) {

	fmt.Printf("got from db (%d)\n", request.Nonce)

	if common.GetConfigInstance().Debug {
		messengerThread.logger.Println("[etl_messenger] received ping over C3")
	}

	fmt.Printf("send to db (%d, %t)\n", request.Nonce, true)
	messengerThread.C4 <- threads.MessengerResponse{Nonce: request.Nonce, Success: true}
}

func (messengerThread *Thread) ProcessConsoleRequest(request *threads.MessengerRequest) {
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

func (messengerThread *Thread) ProcessCloseLogRequest(request *threads.MessengerRequest) {
	messengerInstance := GetMessengerInstance()
	messengerInstance.Complete(request.Cluster)
}

func (messengerThread *Thread) Teardown() {
	messengerThread.accepting = false

	messengerThread.wg.Wait()
}
