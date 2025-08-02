package socket

import (
	"github.com/GabeCordo/Flock/internal/shared/logging"
	"github.com/GabeCordo/Flock/internal/shared/socket"
)

const privateKeyPathEnvVar = "FLOCK_SERVER_TLS_KEY"
const certificatePathEnvVar = "FLOCK_SERVER_TLS_CERT"

type UseCases struct {
	Socket socket.Server
	Logger logging.Logger
}
