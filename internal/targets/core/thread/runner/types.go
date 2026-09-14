package runner

import (
	"errors"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/nonce"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/thread"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/use_cases/runner"
)

type Config struct {
	Debug   bool
	Timeout float64
}

type Thread struct {
	config   *Config
	channels struct {
		interrupt chan thread.InterruptEvent

		c13 chan *thread.Request  // runner receives requests from the processor
		c14 chan *thread.Response // runner sends responses to the processor

		c15 chan *thread.Request  //runner sends requests to the database
		c16 chan *thread.Response // runner receives responses from the database

		c9  chan *thread.Request  // runner sends requests to the tls-socket
		c10 chan *thread.Response // runner receives responses from the tls-socket

		c17 chan *thread.Request // runner sends requests to the messenger

		close chan thread.InterruptEvent
	}
	useCases     runner.UseCases
	logger       logging.Logger
	requestStore map[nonce.Nonce]*thread.Request
}

func NewThread(cfg *Config, logger logging.Logger, useCases runner.UseCases, channels ...any) (*Thread, error) {
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
	t.channels.c13, ok = (channels[1]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorRequest' in index 1")
	}
	t.channels.c14, ok = (channels[2]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan SupervisorResponse' in index 2")
	}
	t.channels.c15, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 3")
	}
	t.channels.c16, ok = (channels[4]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 4")
	}
	t.channels.c17, ok = (channels[5]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 5")
	}
	t.channels.c9, ok = (channels[6]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan RunnerRequest' in index 6")
	}
	t.channels.c10, ok = (channels[7]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan RunnerResponse' in index 7")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	t.requestStore = make(map[nonce.Nonce]*thread.Request)

	t.useCases = useCases

	return t, nil
}
