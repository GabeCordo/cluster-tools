package cache

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/shared/logging"
	"github.com/FortifiedCode/flock/internal/shared/terminal"
	"github.com/FortifiedCode/flock/internal/targets/core/component/cache"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
	cache2 "github.com/FortifiedCode/flock/internal/targets/core/use_cases/cache"
)

type Config struct {
	Debug bool
}

type Thread struct {
	config   *Config
	channels struct {
		interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

		c9  <-chan *thread.Request  // cache receiving requests from the rest processor
		c10 chan<- *thread.Response // cache sending responses to the rest processor

		c24 <-chan *thread.Request  // cache receiving requests from the rest thread
		c25 chan<- *thread.Response // cache sending responses to the rest rest

		close chan thread.InterruptEvent
	}
	useCase cache2.UseCase
	logger  logging.Logger
}

func New(cfg *Config, logger logging.Logger, cache cache.Cache, channels ...any) (*Thread, error) {
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
	t.channels.c9, ok = (channels[1]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 1")
	}
	t.channels.c10, ok = (channels[2]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 2")
	}
	t.channels.c24, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan CacheRequest' in index 3")
	}
	t.channels.c25, ok = (channels[4]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan CacheResponse' in index 4")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	t.logger = logger
	t.logger.SetColour(terminal.Yellow)

	t.useCase = cache2.UseCase{
		CacheComponent: cache,
	}

	return t, nil
}
