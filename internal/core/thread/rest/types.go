package rest

import (
	"context"
	"errors"
	"github.com/GabeCordo/Flock/internal/core/thread"
	nonce2 "github.com/GabeCordo/Flock/internal/shared/nonce"
	"github.com/GabeCordo/toolchain/logging"
	"net/http"
)

// Frontend Thread

type Config struct {
	Debug      bool
	EnableCors bool
	Net        struct {
		Host string
		Port int
	}
	Timeout float64
}

type Thread struct {
	config   *Config
	channels struct {
		interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

		c1 chan<- *thread.Request  // Core is sending thread to the Database
		c2 <-chan *thread.Response // Core is receiving responses from the Database

		c5 chan<- *thread.Request  // Core is sending thread to the Database
		c6 <-chan *thread.Response // Core is receiving responses from the Database

		c20 chan<- *thread.Request  // ClientHttp is sending requests to the Scheduler
		c21 <-chan *thread.Response // ClientHttp is receiving responses from the Scheduler

		c22 chan<- *thread.Request  // Core is sending requests to the Messenger
		c23 <-chan *thread.Response // Core is receiving responses from the Messenger

		close chan thread.InterruptEvent
	}
	logger                 *logging.Logger
	noncePool              *nonce2.Pool
	ProcessorResponseTable *nonce2.ResponseTable
	DatabaseResponseTable  *nonce2.ResponseTable
	SchedulerResponseTable *nonce2.ResponseTable
	MessengerResponseTable *nonce2.ResponseTable
	CacheResponseTable     *nonce2.ResponseTable
	server                 *http.Server
	mux                    *http.ServeMux
	cancelCtx              context.CancelFunc
}

func New(cfg *Config, logger *logging.Logger, noncePool *nonce2.Pool, channels ...any) (*Thread, error) {
	t := new(Thread)

	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	t.channels.interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	t.channels.c1, ok = (channels[1]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	t.channels.c2, ok = (channels[2]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	t.channels.c5, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 3")
	}
	t.channels.c6, ok = (channels[4]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 4")
	}
	t.channels.c20, ok = (channels[5]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorRequest' in index 5")
	}
	t.channels.c21, ok = (channels[6]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan ProcessorResponse' in index 6")
	}
	t.channels.c22, ok = (channels[7]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 7")
	}
	t.channels.c23, ok = (channels[8]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 8")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	t.noncePool = noncePool
	t.ProcessorResponseTable = nonce2.NewResponseTable()
	t.DatabaseResponseTable = nonce2.NewResponseTable()
	t.SchedulerResponseTable = nonce2.NewResponseTable()
	t.MessengerResponseTable = nonce2.NewResponseTable()
	t.CacheResponseTable = nonce2.NewResponseTable()

	t.server = new(http.Server)

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	t.logger = logger
	t.logger.SetColour(logging.Green)

	return t, nil
}
