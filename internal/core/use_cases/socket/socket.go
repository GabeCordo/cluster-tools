package socket

import (
	"crypto/tls"
	"net"
	"os"

	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
)

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
		err = uc.Socket.Setup(cert)
		if err != nil {
			return err
		}
	} else {
		uc.Logger.Alertln("the gateway has defaulted to an unencrypted socket! Do NOT use in production!")
		err := uc.Socket.Setup()
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *UseCases) StartNetworkSocket(host string, port int,
	CreateProcessor func(pId uint64, pAddr string) error,
	DeleteProcessor func(pId uint64, pAddr string) error,
	CreateModule func(pId uint64, m *processor2.ModuleConfig),
	UpdateRun func(r *run.Run)) {

	uc.Socket.Start(host, port,
		CreateProcessor, DeleteProcessor, CreateModule, UpdateRun)
}

func (uc *UseCases) GetProcessor(pId uint64) (connection net.Conn, err error) {

	return uc.Socket.GetConnection(pId)
}
