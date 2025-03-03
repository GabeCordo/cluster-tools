package socket

import (
	"crypto/tls"
	"net"
	"time"
)

func (thread *Thread) createSocketConnection(gatewayHost string) (connection net.Conn, err error) {

	for i := 0; i < MaxNumberOfRetries; i++ {

		if thread.flags.useTLS {
			config := &tls.Config{RootCAs: thread.tls.pool}
			connection, err = tls.Dial("tcp", gatewayHost, config)
		} else {
			connection, err = net.Dial("tcp", gatewayHost)
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
