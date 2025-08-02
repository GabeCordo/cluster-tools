package database

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/thread"
	database2 "github.com/GabeCordo/Flock/internal/core/use_cases/database"
	"github.com/GabeCordo/Flock/internal/shared/logging"
	"github.com/GabeCordo/Flock/internal/shared/nonce"
	"github.com/GabeCordo/Flock/internal/shared/terminal"
)

var StoreTypeMismatch = errors.New("the received type and desired database type do not match")

type Config struct {
	Debug            bool
	Timeout          float64
	Type             string
	ConfigsFolder    string
	StatisticsFolder string
}

type Thread struct {
	config   *Config
	channels struct {
		interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

		c1 <-chan *thread.Request  // Database is receiving thread from the http_thread
		c2 chan<- *thread.Response // Database is sending responses to the http_thread

		c3 chan<- *thread.Request  // Database is sending thread to the Messenger
		c4 <-chan *thread.Response // Database is receiving responses from the Messenger

		c11 <-chan *thread.Request  // Database is receiving req from the processor_thread
		c12 chan<- *thread.Response // Database is sending rsp to the processor_thread

		c15 <-chan *thread.Request  // Database is receiving req from the supervisor_thread
		c16 chan<- *thread.Response // Database is sending rsp from the supervisor_thread

		c26 <-chan *thread.Request  // Database is receiving req from the scheduler_thread
		c27 chan<- *thread.Response // Database is sending rsp to the scheduler_thread

		close chan thread.InterruptEvent
	}
	useCases               database2.UseCases
	logger                 logging.Logger
	messengerResponseTable *nonce.ResponseTable
}

func New(cfg *Config, logger logging.Logger, useCases database2.UseCases, channels ...interface{}) (*Thread, error) {

	t := new(Thread)
	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	t.useCases = useCases

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
	t.channels.c3, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	t.channels.c4, ok = (channels[4]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 4")
	}
	t.channels.c11, ok = (channels[5]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}
	t.channels.c12, ok = (channels[6]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}
	t.channels.c15, ok = (channels[7]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 7")
	}
	t.channels.c16, ok = (channels[8]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 8")
	}
	t.channels.c26, ok = (channels[9]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 9")
	}
	t.channels.c27, ok = (channels[10]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 10")
	}
	t.channels.close = make(chan thread.InterruptEvent)

	t.messengerResponseTable = nonce.NewResponseTable()

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	t.logger = logger
	t.logger.SetColour(terminal.Purple)

	return t, nil
}
