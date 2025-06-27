package socket

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/thread"
	nonce2 "github.com/GabeCordo/Flock/internal/shared/nonce"
	"github.com/GabeCordo/toolchain/logging"
)

const nonceMin = 524228
const nonceMax = 786342 // (base) 524228 + 262114 (offset)

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
	channels struct {
		interrupt chan<- thread.InterruptEvent
		c7        chan<- *thread.Request  // socket_thread is sending req to the processor_thread
		c8        <-chan *thread.Response // socket_thread is rec rsp from the processor_thread
		c9        <-chan *thread.Request  // runner_thread is sending req to the socket_thread
		c10       chan<- *thread.Response // socket_thread is sending rsp to the runner_thread
		close     chan thread.InterruptEvent
	}

	noncePool *nonce2.Pool

	responseTables struct {
		processor *nonce2.ResponseTable
		runner    *nonce2.ResponseTable
	}

	flags struct {
		useTLS bool
	}

	tls struct {
		config *tls.Config
	}

	connections      map[uint64]net.Conn
	numOfConnections uint64

	config *Config
	Logger *logging.Logger

	wg             sync.WaitGroup
	connectionsMux sync.RWMutex
	mutex          sync.RWMutex
}

func New(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {
	t := new(Thread)

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	if logger != nil {
		t.Logger = logger
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

	t.connections = make(map[uint64]net.Conn)

	t.noncePool = nonce2.New(nonceMin, nonceMax)

	t.responseTables.processor = nonce2.NewResponseTable()
	t.responseTables.runner = nonce2.NewResponseTable()

	return t, nil
}
