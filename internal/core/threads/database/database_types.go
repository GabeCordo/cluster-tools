package database

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"github.com/GabeCordo/toolchain/multithreaded"
	"sync"
)

type Config struct {
	Debug   bool
	Timeout float64
	Type    string
}

type Thread struct {
	Interrupt chan<- common.InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 <-chan common.ThreadRequest  // Database is receiving threads from the http_thread
	C2 chan<- common.ThreadResponse // Database is sending responses to the http_thread

	C3 chan<- common.ThreadRequest  // Database is sending threads to the Messenger
	C4 <-chan common.ThreadResponse // Database is receiving responses from the Messenger

	C11 <-chan common.ThreadRequest  // Database is receiving req from the processor_thread
	C12 chan<- common.ThreadResponse // Database is sending rsp to the processor_thread

	C15 <-chan common.ThreadRequest  // Database is receiving req from the supervisor_thread
	C16 chan<- common.ThreadResponse // Database is sending rsp from the supervisor_thread

	C26 <-chan common.ThreadRequest  // Database is receiving req from the scheduler_thread
	C27 chan<- common.ThreadResponse // Database is sending rsp to the scheduler_thread

	messengerResponseTable *multithreaded.ResponseTable

	configFolderPath    string
	statisticFolderPath string

	config *Config
	logger *logging.Logger

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, configPath, statisticPath string, channels ...interface{}) (*Thread, error) {
	thread := new(Thread)
	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *config type")
	}
	thread.config = cfg

	thread.configFolderPath = configPath
	thread.statisticFolderPath = statisticPath

	thread.Interrupt, ok = (channels[0]).(chan common.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	thread.C1, ok = (channels[1]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 1")
	}
	thread.C2, ok = (channels[2]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 2")
	}
	thread.C3, ok = (channels[3]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	thread.C4, ok = (channels[4]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 4")
	}
	thread.C11, ok = (channels[5]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 5")
	}
	thread.C12, ok = (channels[6]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 6")
	}
	thread.C15, ok = (channels[7]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 7")
	}
	thread.C16, ok = (channels[8]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 8")
	}
	thread.C26, ok = (channels[9]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseRequest' in index 9")
	}
	thread.C27, ok = (channels[10]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan DatabaseResponse' in index 10")
	}

	thread.messengerResponseTable = multithreaded.NewResponseTable()

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	thread.logger = logger
	thread.logger.SetColour(logging.Purple)

	return thread, nil
}
