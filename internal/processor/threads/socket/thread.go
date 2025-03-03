package socket

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	common "github.com/Sentmint/pops/internal"
	"github.com/Sentmint/pops/internal/core/database/run"
	"github.com/Sentmint/pops/internal/processor/threads"
	"github.com/Sentmint/yule"
	"log"
	"os"
)

func (thread *Thread) Setup() {

	certificatePath := os.Getenv("CTOOLS_TLS_CERT")
	if certificatePath == "" {
		thread.logger.Warnln("CTOOLS_TLS_CERT environment variable not set")
	}

	// [requirements]
	// 1. the processor shall use an un-encrypted socket when the tls-certificate is missing
	// 2. the processor shall output a console warn when an un-encrypted socket is opened
	if certificatePath == "" {
		thread.logger.Alertln("the processor has opened an unencrypted socket to the gateway! Do NOT use in production!")
		thread.flags.useTLS = false
	} else {
		thread.flags.useTLS = true
	}

	if thread.flags.useTLS {
		cert, err := os.ReadFile(certificatePath)
		if err != nil {
			panic(err)
		}

		thread.tls.pool = x509.NewCertPool()
		if ok := thread.tls.pool.AppendCertsFromPEM(cert); !ok {
			log.Fatalf("failed to parse root certificate from %s",
				thread.Config.Tls.Certificate)
		}
	}

	// CONNECTION TO GATEWAY

	var gatewayHost string

	envGatewayHost := os.Getenv("CTOOLS_GATEWAY_HOST")
	if envGatewayHost == "" {
		gatewayHost = *thread.Config.Core
	} else {
		gatewayHost = envGatewayHost
	}

	connection, err := thread.createSocketConnection(gatewayHost)
	if err != nil {
		panic(err)
	}

	// todo : is this the best practice?
	thread.connection = connection
}

func (thread *Thread) Start() {

	// INCOMING REQUESTS

	go func() {
		for request := range thread.channels.C0 {
			if !thread.accepting {
				break
			}
			thread.requestWg.Add(1)
			thread.ProcessRequest(&request)
			thread.requestWg.Done()
		}
	}()

	// RESPONSE THREADS

	go func() {
		for response := range thread.channels.C2 {
			if !thread.accepting {
				break
			}
			thread.ProvisionerResponseTable.Write(response.Nonce, response)
		}
	}()

	// LISTEN TO GATEWAY

	decoder := json.NewDecoder(thread.connection)

	data := &common.Request{}

	for {
		err := decoder.Decode(data)
		if err != nil {

			// non-nil error indicates the socket was closed between the core-processor
			// we need to attempt a reconnect to avoid interruptions
			thread.logger.Alertln("gateway sent EOF closing the socket connection")

			connection, err := thread.createSocketConnection("")
			if err != nil {
				// the core could have shutdown or crashed
				break
			}
			thread.connection = connection

			// listen to the next request coming from the socket
			continue
		}
		thread.ProcessSocketRequest(data)
	}

	thread.connection.Close()
	thread.channels.Interrupt <- threads.Shutdown
}

func (thread *Thread) ProcessRequest(request *threads.SocketRequest) {

	if thread.connection == nil {
		thread.logger.Alertln("cannot process request because connection is nil")
		return
	}

	switch request.Action {
	case threads.SocketModuleAdd:
		{
			module, ok := request.Data.(yule.Module)
			if !ok {
				thread.logger.Warnln("received module add with invalid data")
				return
			}

			req := &common.Request{
				Action: common.Create,
				Record: common.Module,
				Data:   module,
			}

			encoder := json.NewEncoder(thread.connection)
			err := encoder.Encode(req)
			if err != nil {
				fmt.Println(err)
				thread.logger.Warnln("failed to add module over socket")
			}
		}
	case threads.SocketRunUpdate:
		{
			run, ok := request.Data.(run.Run)
			if !ok {
				thread.logger.Warnln("received run update with invalid data")
			}

			req := &common.Request{
				Action: common.Update,
				Record: common.Run,
				Data:   run,
			}

			encoder := json.NewEncoder(thread.connection)
			err := encoder.Encode(req) // todo : fix
			if err != nil {
				fmt.Println(err)
				thread.logger.Warnln("failed to add module over socket")
			}
		}
	default:
		{
			thread.logger.Warnln("received invalid socket request action")
		}
	}
}

func (thread *Thread) ProcessSocketRequest(request *common.Request) {

	switch request.Action {
	case common.Create:
		{
			switch request.Record {
			case common.Run:
				{
					b, err := json.Marshal(request.Data)
					if err != nil {
						thread.logger.Warnln("failed to marshal the received data")
						return
					}

					runRequest := new(run.Request)
					err = json.Unmarshal(b, runRequest)
					if err != nil {
						thread.logger.Warnln("received invalid data for update run")
						return
					}

					err = threads.RunProvision(thread.channels.C1, thread.ProvisionerResponseTable,
						runRequest.Namespace, runRequest.Id, runRequest.Config, runRequest.Metadata, *thread.Config.Timeout)
					if err != nil {
						thread.logger.Warnln("failed to send run provision to provisioner")
					}
				}
			default:
				{
					thread.logger.Warnln("received invalid socket create record")
				}
			}
		}
	case common.Delete:
		{
			switch request.Record {
			case common.Run:
				{
					id, ok := request.Data.(float64)
					if !ok {
						thread.logger.Warnln("run delete received value other than uint64")
						return
					}

					err := threads.RunStop(thread.channels.C1, thread.ProvisionerResponseTable,
						uint64(id), *thread.Config.Timeout)
					if err != nil {
						thread.logger.Warnln("failed to stop ongoing run")
					}
				}
			default:
				{
					thread.logger.Warnln("received invalid socket delete record")
				}
			}
		}
	default:
		{
			thread.logger.Warnln("received invalid socket request action")
		}
	}
}

func (thread *Thread) Teardown() {

	if thread.connection != nil {
		thread.connection.Close()
	}
	thread.requestWg.Wait()
}
