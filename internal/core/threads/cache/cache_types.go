package cache

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"sync"
)

type Config struct {
	Debug bool
}

type Thread struct {
	Interrupt chan<- common.InterruptEvent // Upon completion or failure an interrupt can be raised

	C9  <-chan common.ThreadRequest  // cache receiving requests from the http processor
	C10 chan<- common.ThreadResponse // cache sending responses to the http processor

	C24 <-chan common.ThreadRequest  // cache receiving requests from the http client
	C25 chan<- common.ThreadResponse // cache sending responses to the http client

	config *Config
	logger *logging.Logger

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, channels ...any) (*Thread, error) {
	thread := new(Thread)
	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *config type")
	}
	thread.config = cfg

	thread.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	thread.C9, ok = (channels[1]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 1")
	}
	thread.C10, ok = (channels[2]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 2")
	}
	thread.C24, ok = (channels[3]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 3")
	}
	thread.C25, ok = (channels[4]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 4")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	thread.logger = logger
	thread.logger.SetColour(logging.Yellow)

	return thread, nil
}
