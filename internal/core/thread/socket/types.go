package socket

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/Flock/internal/core/use_cases/socket"
	nonce2 "github.com/GabeCordo/Flock/internal/shared/nonce"
	socket2 "github.com/GabeCordo/Flock/internal/shared/socket"
	"github.com/GabeCordo/toolchain/logging"
)

type Config struct {
	Debug bool
	Net   struct {
		Host string
		Port int
	}
	Tls struct {
		Certificate string
		Key         string
	}
	Timeout float64
}

type Thread struct {
	config   *Config
	channels struct {
		interrupt chan<- thread.InterruptEvent
		c7        chan<- *thread.Request  // socket_thread is sending req to the processor_thread
		c8        <-chan *thread.Response // socket_thread is rec rsp from the processor_thread
		c9        <-chan *thread.Request  // runner_thread is sending req to the socket_thread
		c10       chan<- *thread.Response // socket_thread is sending rsp to the runner_thread
		close     chan thread.InterruptEvent
	}
	useCases  *socket.UseCases
	logger    *logging.Logger
	noncePool *nonce2.Pool

	serverSocket socket2.Server

	responseTables struct {
		processor *nonce2.ResponseTable
		runner    *nonce2.ResponseTable
	}
}

func New(cfg *Config, logger *logging.Logger, noncePool *nonce2.Pool, useCases *socket.UseCases, channels ...any) (*Thread, error) {
	t := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	if logger != nil {
		t.logger = logger
	} else {
		return nil, errors.New("expected logger to be a non-nil value")
	}

	var ok = false

	t.channels.interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}

	t.channels.c7, ok = (channels[1]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.channels.c8, ok = (channels[2]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}

	t.channels.c9, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 1")
	}

	t.channels.c10, ok = (channels[4]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 2")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	t.useCases = useCases

	t.noncePool = noncePool

	t.responseTables.processor = nonce2.NewResponseTable()
	t.responseTables.runner = nonce2.NewResponseTable()

	return t, nil
}
