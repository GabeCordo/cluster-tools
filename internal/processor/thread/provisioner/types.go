package provisioner

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	"github.com/GabeCordo/Flock/internal/processor/use_cases/provisioner"
	"github.com/GabeCordo/Flock/internal/shared/logging"
	"github.com/GabeCordo/Flock/internal/shared/nonce"
	"github.com/GabeCordo/Flock/internal/shared/terminal"
)

const nonceMin = 1000000
const nonceMax = 2000000

type Config struct {
	Debug     *bool
	Timeout   *float64
	Core      *string
	Processor processor.Config
}

type Thread struct {
	Config   *Config
	channels struct {
		Interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

		C0 chan *thread.SocketRequest
		C1 chan *thread.ProvisionerRequest    // Runtime is receiving thread from the http_thread
		C2 chan<- *thread.ProvisionerResponse // Runtime is sending responses to the http_thread

		close chan thread.InterruptEvent
	}
	useCases  *provisioner.UseCases
	logger    logging.Logger
	noncePool *nonce.Pool
}

func NewThread(cfg *Config, useCases *provisioner.UseCases, channels ...interface{}) (*Thread, error) {
	t := new(Thread)
	var ok bool

	t.channels.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	t.channels.C0, ok = (channels[1]).(chan *thread.SocketRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SocketRequest' in index 1")
	}
	t.channels.C1, ok = (channels[2]).(chan *thread.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 2")
	}
	t.channels.C2, ok = (channels[3]).(chan *thread.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 3")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	t.useCases = useCases
	t.logger = t.useCases.Logger

	if cfg == nil {
		return nil, errors.New("expected no nil *modules.pipeline type")
	}
	t.Config = cfg

	t.noncePool = nonce.New(nonceMin, nonceMax)

	t.logger.SetColour(terminal.Orange)

	return t, nil
}
