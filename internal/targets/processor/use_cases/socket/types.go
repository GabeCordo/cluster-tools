package socket

import (
	"github.com/GabeCordo/FunctionScheduler/internal/shared/logging"
	"github.com/GabeCordo/FunctionScheduler/internal/shared/socket"
)

const coreHostEnvVar = "FunctionScheduler_CORE_HOST"
const tlsCertEnvVar = "FunctionScheduler_CLIENT_TLS_CERT"

type UseCases struct {
	Sock   socket.Client
	Logger logging.Logger
}
