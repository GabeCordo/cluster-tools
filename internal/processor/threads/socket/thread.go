package socket

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	common "github.com/GabeCordo/Flock/internal"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/processor"
	"github.com/GabeCordo/Flock/internal/processor/threads"
)

func (thread *Thread) attemptConnectionToCore(host string) (connection net.Conn, err error) {

	for i := 0; i < MaxNumberOfRetries; i++ {

		if thread.flags.useTLS {
			config := &tls.Config{RootCAs: thread.tls.pool}
			connection, err = tls.Dial("tcp", host, config)
		} else {
			thread.logger.Warnln("connecting on a non-encrypted channel!")
			connection, err = net.Dial("tcp", host)
		}

		if err != nil {
			thread.logger.Warnf("failed to connect to gateway (retry: %d)\n", i)
		} else {
			thread.logger.Println("connected to gateway")
			break
		}

		time.Sleep(MaxWaitBeforeRetry)
	}

	return connection, err
}

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

	if connection, err := thread.attemptConnectionToCore(gatewayHost); err == nil {
		thread.connection = connection
	} else {
		panic(err)
	}
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

	// Loop
	for {

		// Loop: listen to incoming messages from the core.
		for {
			err := decoder.Decode(data)
			if err != nil {
				thread.logger.Alertln("gateway sent EOF closing the socket connection")
				break
			}
			thread.ProcessSocketRequest(data)
		}

		// TODO: we are violating DRY here, this is a quick hack
		for {
			var host string
			envHost := os.Getenv("CTOOLS_GATEWAY_HOST")

			if envHost == "" {
				host = *thread.Config.Core
			} else {
				host = envHost
			}

			// Condition: connection to the core has dropped, attempt to connect
			if connection, err := thread.attemptConnectionToCore(host); err == nil {
				thread.connection = connection
				decoder = json.NewDecoder(thread.connection)
				break
			} else {
				// Edge-Case: the processor was unable to re-establish connection to the core.
			}
		}
	}

	err := thread.connection.Close()
	if err != nil {
		fmt.Print(err)
	}
	thread.connection = nil
	thread.channels.Interrupt <- threads.Shutdown
}

func (thread *Thread) ProcessRequest(request *threads.SocketRequest) {

	// Edge Case: it is possible that the core drops while the processor is alive
	// Behaviour: the processor shall ignore requests received on its socket until the connection
	// 			  is re-established to the core.
	// TODO : could we buffer messages until the core comes online?
	if thread.connection == nil {
		// Since the socket does not return a response, we cannot do
		// anything until this is enhanced.
		return
	}

	switch request.Action {
	case threads.SocketModuleAdd:
		{
			module, ok := request.Data.(processor.ModuleConfig)
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
			r, ok := request.Data.(run.Run)
			if !ok {
				thread.logger.Warnln("received run update with invalid data")
			}

			req := &common.Request{
				Action: common.Update,
				Record: common.Run,
				Data:   r,
			}

			encoder := json.NewEncoder(thread.connection)
			err := encoder.Encode(req) // todo : fix
			if err != nil {
				fmt.Println(err)
				thread.logger.Warnln("failed to update run over socket")
			}
		}
	case threads.SocketLogAdd:
		{

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

	if thread.connection == nil {
		return
	}

	err := thread.connection.Close()
	if err != nil {
		fmt.Print(err)
	}
	thread.requestWg.Wait()
}
