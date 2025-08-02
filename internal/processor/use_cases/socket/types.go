package socket

import (
	"github.com/GabeCordo/Flock/internal/shared/logging"
	"github.com/GabeCordo/Flock/internal/shared/socket"
)

const coreHostEnvVar = "FLOCK_CORE_HOST"
const tlsCertEnvVar = "FLOCK_CLIENT_TLS_CERT"

type UseCases struct {
	Sock   socket.Client
	Logger logging.Logger
}
