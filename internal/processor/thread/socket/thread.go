package socket

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	common "github.com/GabeCordo/Flock/internal/shared/async"
)

func (t *Thread) attemptConnectionToCore(host string) (connection net.Conn, err error) {

	for i := 0; i < MaxNumberOfRetries; i++ {

		if t.flags.useTLS {
			config := &tls.Config{
				RootCAs:    t.tls.pool,
				MinVersion: tls.VersionTLS12,
			}
			connection, err = tls.Dial("tcp", host, config)
		} else {
			t.logger.Warnln("connecting on a non-encrypted channel!")
			connection, err = net.Dial("tcp", host)
		}

		if err != nil {
			t.logger.Warnf("failed to connect to gateway (retry: %d)\n", i)
		} else {
			t.logger.Println("connected to gateway")
			break
		}

		time.Sleep(MaxWaitBeforeRetry)
	}

	return connection, err
}

func (t *Thread) Setup() {

	certificatePath := os.Getenv("CTOOLS_TLS_CERT")
	if certificatePath == "" {
		t.logger.Warnln("CTOOLS_TLS_CERT environment variable not set")
	}

	// [requirements]
	// 1. the processor shall use an un-encrypted socket when the tls-certificate is missing
	// 2. the processor shall output a console warn when an un-encrypted socket is opened
	if certificatePath == "" {
		t.logger.Alertln("the processor has opened an unencrypted socket to the gateway! Do NOT use in production!")
		t.flags.useTLS = false
	} else {
		t.flags.useTLS = true
	}

	// clean the file path provided to the program in CTOOLS_TLS_CERT
	certificatePath = filepath.Clean(certificatePath)

	if t.flags.useTLS {
		cert, err := os.ReadFile(certificatePath)
		if err != nil {
			panic(err)
		}

		t.tls.pool = x509.NewCertPool()
		if ok := t.tls.pool.AppendCertsFromPEM(cert); !ok {
			log.Fatalf("failed to parse root certificate from %s",
				t.Config.Tls.Certificate)
		}
	}

	// CONNECTION TO GATEWAY

	var gatewayHost string
	envGatewayHost := os.Getenv("CTOOLS_GATEWAY_HOST")

	if envGatewayHost == "" {
		gatewayHost = *t.Config.Core
	} else {
		gatewayHost = envGatewayHost
	}

	if connection, err := t.attemptConnectionToCore(gatewayHost); err == nil {
		t.connection = connection
	} else {
		panic(err)
	}
}

func (t *Thread) Start() {

	// INCOMING REQUESTS

	go func() {
		for request := range t.channels.C0 {
			if !t.accepting {
				break
			}
			t.requestWg.Add(1)
			t.ProcessRequest(&request)
			t.requestWg.Done()
		}
	}()

	// RESPONSE THREADS

	go func() {
		for response := range t.channels.C2 {
			if !t.accepting {
				break
			}
			t.ProvisionerResponseTable.Write(response.Nonce, response)
		}
	}()

	// LISTEN TO GATEWAY

	decoder := json.NewDecoder(t.connection)

	data := &common.Request{}

	// Loop
	for {

		// Loop: listen to incoming messages from the core.
		for {
			err := decoder.Decode(data)
			if err != nil {
				t.logger.Alertln("gateway sent EOF closing the socket connection")
				break
			}
			t.ProcessSocketRequest(data)
		}

		// TODO: we are violating DRY here, this is a quick hack
		for {
			var host string
			envHost := os.Getenv("CTOOLS_GATEWAY_HOST")

			if envHost == "" {
				host = *t.Config.Core
			} else {
				host = envHost
			}

			// Condition: connection to the core has dropped, attempt to connect
			if connection, err := t.attemptConnectionToCore(host); err == nil {
				t.connection = connection
				decoder = json.NewDecoder(t.connection)
				break
			} else {
				// Edge-Case: the processor was unable to re-establish connection to the core.
			}
		}
	}

	err := t.connection.Close()
	if err != nil {
		fmt.Print(err)
	}
	t.connection = nil
	t.channels.Interrupt <- thread.Shutdown
}

func (t *Thread) ProcessRequest(request *thread.SocketRequest) {

	// Edge Case: it is possible that the core drops while the processor is alive
	// Behaviour: the processor shall ignore requests received on its socket until the connection
	// 			  is re-established to the core.
	// TODO : could we buffer messages until the core comes online?
	if t.connection == nil {
		// Since the socket does not return a response, we cannot do
		// anything until this is enhanced.
		return
	}

	switch request.Action {
	case thread.SocketModuleAdd:
		{
			module, ok := request.Data.(processor.ModuleConfig)
			if !ok {
				t.logger.Warnln("received module add with invalid data")
				return
			}

			req := &common.Request{
				Action: common.Create,
				Record: common.Module,
				Data:   module,
			}

			encoder := json.NewEncoder(t.connection)
			err := encoder.Encode(req)
			if err != nil {
				fmt.Println(err)
				t.logger.Warnln("failed to add module over socket")
			}
		}
	case thread.SocketRunUpdate:
		{
			r, ok := request.Data.(run.Run)
			if !ok {
				t.logger.Warnln("received run update with invalid data")
			}

			req := &common.Request{
				Action: common.Update,
				Record: common.Run,
				Data:   r,
			}

			encoder := json.NewEncoder(t.connection)
			err := encoder.Encode(req) // todo : fix
			if err != nil {
				fmt.Println(err)
				t.logger.Warnln("failed to update run over socket")
			}
		}
	case thread.SocketLogAdd:
		{

		}
	default:
		{
			t.logger.Warnln("received invalid socket request action")
		}
	}
}

func (t *Thread) ProcessSocketRequest(request *common.Request) {

	switch request.Action {
	case common.Create:
		{
			switch request.Record {
			case common.Run:
				{
					b, err := json.Marshal(request.Data)
					if err != nil {
						t.logger.Warnln("failed to marshal the received data")
						return
					}

					runRequest := new(run.Request)
					err = json.Unmarshal(b, runRequest)
					if err != nil {
						t.logger.Warnln("received invalid data for update run")
						return
					}

					mandatory := thread.ProvisionerMandatory{
						Pipe:          t.channels.C1,
						ResponseTable: t.ProvisionerResponseTable,
						NoncePool:     t.noncePool,
						Timeout:       *t.Config.Timeout,
					}
					err = thread.RunStart(mandatory,
						runRequest.Namespace, runRequest.Id, runRequest.Config, runRequest.Metadata)
					if err != nil {
						t.logger.Warnln("failed to send run provision to provisioner")
					}
				}
			default:
				{
					t.logger.Warnln("received invalid socket create record")
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
						t.logger.Warnln("run delete received value other than uint64")
						return
					}

					mandatory := thread.ProvisionerMandatory{
						Pipe:          t.channels.C1,
						ResponseTable: t.ProvisionerResponseTable,
						NoncePool:     t.noncePool,
						Timeout:       *t.Config.Timeout,
					}
					err := thread.RunStop(mandatory, uint64(id))
					if err != nil {
						t.logger.Warnln("failed to stop ongoing run")
					}
				}
			default:
				{
					t.logger.Warnln("received invalid socket delete record")
				}
			}
		}
	default:
		{
			t.logger.Warnln("received invalid socket request action")
		}
	}
}

func (t *Thread) Teardown() {

	if t.connection == nil {
		return
	}

	err := t.connection.Close()
	if err != nil {
		fmt.Print(err)
	}
	t.requestWg.Wait()
}
