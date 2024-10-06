package cache

import (
	"errors"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/Sentmint/cluster-tools/internal/core/cache"
	"github.com/Sentmint/cluster-tools/internal/core/thread"
	"sync"
)

type Config struct {
	Debug bool
}

type Thread struct {
	Interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

	C9  <-chan thread.Request  // cache receiving requests from the rest processor
	C10 chan<- thread.Response // cache sending responses to the rest processor

	C24 <-chan thread.Request  // cache receiving requests from the rest client
	C25 chan<- thread.Response // cache sending responses to the rest client

	config *Config
	logger *logging.Logger

	cache cache.Cache

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, cache cache.Cache, channels ...any) (*Thread, error) {
	t := new(Thread)
	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	t.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	t.C9, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 1")
	}
	t.C10, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 2")
	}
	t.C24, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 3")
	}
	t.C25, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 4")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	t.logger = logger
	t.logger.SetColour(logging.Yellow)

	t.cache = cache

	return t, nil
}
