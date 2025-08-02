package socket

import (
	"crypto/x509"
	"errors"
	"github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/processor/thread"
	"github.com/GabeCordo/Flock/internal/processor/use_cases/socket"
	"github.com/GabeCordo/Flock/internal/shared/logging"
	nonce2 "github.com/GabeCordo/Flock/internal/shared/nonce"
	"github.com/GabeCordo/Flock/internal/shared/terminal"
	"sync"
)

// Frontend Thread

const nonceMin = 0
const nonceMax = 1000000

type Config struct {
	Debug       *bool
	Timeout     *float64
	Standalone  *bool
	ExternalNet processor.Config

	Tls struct {
		Certificate string
	}

	Core *string
	Net  string
}

type Thread struct {
	Config *Config

	channels struct {
		Interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised
		C0        <-chan *thread.SocketRequest
		C1        chan<- *thread.ProvisionerRequest  // Core is sending thread to the Database
		C2        <-chan *thread.ProvisionerResponse // Core is receiving responses from the Database
		close     chan thread.InterruptEvent
	}

	flags struct {
		useTLS bool
	}

	noncePool                *nonce2.Pool
	ProvisionerResponseTable *nonce2.ResponseTable

	tls struct {
		pool *x509.CertPool
	}

	useCases *socket.UseCases

	logger logging.Logger

	requestWg sync.WaitGroup

	counter uint32
	mutex   sync.RWMutex
	wg      sync.WaitGroup
}

func NewThread(cfg *Config, logger logging.Logger, useCases *socket.UseCases, channels ...interface{}) (*Thread, error) {
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
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	t.channels.C2, ok = (channels[3]).(chan *thread.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	t.counter = 0

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	t.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *http.pipeline type")
	}
	t.Config = cfg

	t.noncePool = nonce2.New(nonceMin, nonceMax)

	t.useCases = useCases

	t.logger.SetColour(terminal.Green)

	return t, nil
}
