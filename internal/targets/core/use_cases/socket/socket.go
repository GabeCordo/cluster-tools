package socket

import (
	"crypto/tls"
	socket2 "github.com/FortifiedCode/flock/internal/shared/socket"
	"os"
)

func (uc *UseCases) SetupSocketEventHandlers(events socket2.ServerEvents) {

	uc.Socket.SetupEvents(events)
}

func (uc *UseCases) SetupSocketTlsConfig() error {

	certificatePath := os.Getenv(certificatePathEnvVar)
	if certificatePath == "" {
		uc.Logger.Warnf("%s environment variable not set\n", certificatePathEnvVar)
	}

	keyPath := os.Getenv(privateKeyPathEnvVar)
	if keyPath == "" {
		uc.Logger.Warnf("%s environment variable not set\n", privateKeyPathEnvVar)
	}

	// [requirements]
	// 1. the core(gateway) shall default to an un-encrypted socket when the TLS cert is missing
	// 2. the core(gateway) shall default to an un-encrypted socket when the TLS key is missing
	// 3. the core shall output a warning message when an un-encrypted socket is opened
	useTLS := (certificatePath != "") && (keyPath != "")

	if useTLS {
		cert, err := tls.LoadX509KeyPair(certificatePath, keyPath)
		if err != nil {
			return err
		}
		err = uc.Socket.SetupTLS(cert)
		if err != nil {
			return err
		}
	} else {
		uc.Logger.Alertln("the gateway has defaulted to an unencrypted socket! Do NOT use in production!")
	}

	return nil
}

func (uc *UseCases) StartNetworkSocket(host string, port int) {

	err := uc.Socket.Listen(host, port)
	if err != nil {
		uc.Logger.Alertf("The TCP socket cannot bind to the port %d", port)
		uc.Logger.Alertf("You need to check what service is using port %d", port)
	}
}

func (uc *UseCases) SendDataOnSocket(id socket2.ConnectionId, message *socket2.Message) error {

	return uc.Socket.Send(id, message)
}
