package database

import (
	"errors"
	"sync"

	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
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
	Interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 <-chan thread.Request  // Database is receiving thread from the http_thread
	C2 chan<- thread.Response // Database is sending responses to the http_thread

	C3 chan<- thread.Request  // Database is sending thread to the Messenger
	C4 <-chan thread.Response // Database is receiving responses from the Messenger

	C11 <-chan thread.Request  // Database is receiving req from the processor_thread
	C12 chan<- thread.Response // Database is sending rsp to the processor_thread

	C15 <-chan thread.Request  // Database is receiving req from the supervisor_thread
	C16 chan<- thread.Response // Database is sending rsp from the supervisor_thread

	C26 <-chan thread.Request  // Database is receiving req from the scheduler_thread
	C27 chan<- thread.Response // Database is sending rsp to the scheduler_thread

	messengerResponseTable *multithreaded.ResponseTable

	statisticDatabase database.Database
	pipelineDatabase  database.Database
	jobDatabase       database.Database

	config *Config
	logger *logging.Logger

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger,
	s, c, j database.Database,
	configPath, statisticPath string, channels ...interface{}) (*Thread, error) {

	t := new(Thread)
	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	t.config = cfg

	t.statisticDatabase = s
	t.pipelineDatabase = c
	t.jobDatabase = j

	t.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	t.C1, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	t.C2, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	t.C3, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	t.C4, ok = (channels[4]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 4")
	}
	t.C11, ok = (channels[5]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}
	t.C12, ok = (channels[6]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}
	t.C15, ok = (channels[7]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 7")
	}
	t.C16, ok = (channels[8]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 8")
	}
	t.C26, ok = (channels[9]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 9")
	}
	t.C27, ok = (channels[10]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 10")
	}

	t.messengerResponseTable = multithreaded.NewResponseTable()

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	t.logger = logger
	t.logger.SetColour(logging.Purple)

	return t, nil
}
