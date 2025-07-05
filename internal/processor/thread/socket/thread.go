package socket

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

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

	go t.listenOnSocket()

	var iReq *thread.SocketRequest
	var iRsp *thread.ProvisionerResponse
	stop := false

	for {
		select {
		case iReq = <-t.channels.C0:
			{
				t.requestWg.Add(1)
				t.handleRequest(iReq)
				t.requestWg.Done()
			}
		case iRsp = <-t.channels.C2:
			{
				t.handleResponse(iRsp)
			}
		case <-t.channels.close:
			{
				stop = true
			}
		}

		if stop {
			break
		}
	}
}

func (t *Thread) listenOnSocket() {

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
			t.handleSocketRequest(data)
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
}

func (t *Thread) handleRequest(request *thread.SocketRequest) {

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
			t.handleSocketModuleAdd(request)
		}
	case thread.SocketRunUpdate:
		{
			t.handleSocketRunUpdate(request)
		}
	case thread.SocketLogAdd:
		{
			t.logger.Warnln("socket log add not implemented")
		}
	default:
		{
			t.logger.Warnln("received invalid socket request action")
		}
	}
}

func (t *Thread) handleResponse(response *thread.ProvisionerResponse) {

	if response == nil {
		t.logger.Warnln("received nil response")
		return
	}

	if response.Error != nil {
		t.logger.Warnf("received error from provisioner %s\n", response.Error)
	}
}

func (t *Thread) handleSocketRequest(request *common.Request) {

	switch request.Action {
	case common.Create:
		{
			switch request.Record {
			case common.Run:
				{
					t.handleSocketRunCreate(request)
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
					t.handleSocketRunDelete(request)
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

	if t.connection != nil {
		// close the socket connection with the core
		err := t.connection.Close()
		if err != nil {
			t.logger.Warnf("%s\n", err.Error())
		}
	}

	t.requestWg.Wait()
}
