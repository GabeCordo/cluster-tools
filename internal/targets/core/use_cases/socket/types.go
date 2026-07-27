package socket

import (
	"github.com/GabeCordo/FunctionScheduler/internal/shared/logging"
	"github.com/GabeCordo/FunctionScheduler/internal/shared/socket"
)

const privateKeyPathEnvVar = "FunctionScheduler_SERVER_TLS_KEY"
const certificatePathEnvVar = "FunctionScheduler_SERVER_TLS_CERT"

type UseCases struct {
	Socket socket.Server
	Logger logging.Logger
}
