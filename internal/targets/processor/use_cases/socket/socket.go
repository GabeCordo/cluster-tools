package socket

import (
	"crypto/x509"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/socket"
)

func (uc *UseCases) SetupSocketEventHandlers(events socket.ClientEvents) {

	uc.Sock.SetupEvents(events)
}

func (uc *UseCases) SetupSocketTlsConfig() error {

	certificatePath := os.Getenv(tlsCertEnvVar)
	if certificatePath == "" {
		uc.Logger.Warnf("%s environment variable not set", tlsCertEnvVar)
	}

	// [requirements]
	// 1. the processor shall use an un-encrypted socket when the tls-certificate is missing
	// 2. the processor shall output a console warn when an un-encrypted socket is opened
	if certificatePath == "" {
		uc.Logger.Alertln("the processor has opened an unencrypted socket to the gateway! Do NOT use in production!")
		return nil
	}

	// clean the file path provided to the program in CTOOLS_TLS_CERT
	certificatePath = filepath.Clean(certificatePath)

	cert, err := os.ReadFile(certificatePath)
	if err != nil {
		return err
	}

	TLSPool := x509.NewCertPool()
	if ok := TLSPool.AppendCertsFromPEM(cert); !ok {
		log.Fatalf("failed to parse root certificate from %s", certificatePath)
	}

	err = uc.Sock.SetupTLS(TLSPool)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCases) ConnectToServer(host string) error {

	var coreHost string
	envCoreHost := os.Getenv(coreHostEnvVar)

	if (envCoreHost == "") && (host == "") {
		return errors.New("cannot have the environment variable and config host as nil")
	}

	if envCoreHost == "" {
		coreHost = host
	} else {
		coreHost = envCoreHost
	}

	err := uc.Sock.Connect(coreHost)
	if err != nil {
		uc.Logger.Alertf("failed to connect to server %s", host)
	}

	return nil
}

func (uc *UseCases) IsConnectedToCore() bool {

	return uc.Sock.IsConnected()
}

func (uc *UseCases) DisconnectFromCore() {

	err := uc.Sock.Disconnect()
	if err != nil {
		uc.Logger.Alertf("failed to disconnect from core")
	}
}

func (uc *UseCases) SendToCore(message *socket.Message) {

	err := uc.Sock.Send(message)
	if err != nil {
		uc.Logger.Warnln("failed to send message to core")
	}
}
