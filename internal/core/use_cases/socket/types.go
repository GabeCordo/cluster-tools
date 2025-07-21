package socket

import (
	"github.com/GabeCordo/Flock/internal/core/component/net_socket"
	"github.com/GabeCordo/toolchain/logging"
)

const privateKeyPathEnvVar = "CTOOLS_TLS_KEY"
const certificatePathEnvVar = "CTOOLS_TLS_CERT"

type UseCases struct {
	Socket net_socket.NetworkSocket
	Logger *logging.Logger
}
