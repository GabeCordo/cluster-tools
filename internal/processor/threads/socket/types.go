package socket

import (
	"crypto/x509"
	"errors"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"github.com/Sentmint/cluster-tools/internal/core/processor"
	"github.com/Sentmint/cluster-tools/internal/processor/threads"
	"net"
	"sync"
)

// Frontend Thread

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
		Interrupt chan<- threads.InterruptEvent // Upon completion or failure an interrupt can be raised
		C0        <-chan threads.SocketRequest
		C1        chan<- threads.ProvisionerRequest  // Core is sending threads to the Database
		C2        <-chan threads.ProvisionerResponse // Core is receiving responses from the Database
	}

	ProvisionerResponseTable *multithreaded.ResponseTable

	tls struct {
		pool *x509.CertPool
	}

	connection net.Conn

	logger *logging.Logger

	requestWg sync.WaitGroup

	accepting bool
	counter   uint32
	mutex     sync.RWMutex
	wg        sync.WaitGroup
}

func NewThread(cfg *Config, logger *logging.Logger, channels ...interface{}) (*Thread, error) {
	thread := new(Thread)

	var ok bool

	thread.channels.Interrupt, ok = (channels[0]).(chan threads.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	thread.channels.C0, ok = (channels[1]).(chan threads.SocketRequest)
	if !ok {
		return nil, errors.New("expected type 'chan SocketRequest' in index 1")
	}
	thread.channels.C1, ok = (channels[2]).(chan threads.ProvisionerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerRequest' in index 1")
	}
	thread.channels.C2, ok = (channels[3]).(chan threads.ProvisionerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan ProvisionerResponse' in index 2")
	}

	thread.accepting = true
	thread.counter = 0

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	thread.logger = logger

	if cfg == nil {
		return nil, errors.New("expected no nil *http.pipeline type")
	}
	thread.Config = cfg

	thread.ProvisionerResponseTable = multithreaded.NewResponseTable()

	thread.logger.SetColour(logging.Green)

	return thread, nil
}
