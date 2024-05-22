package messenger

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"sync"
)

type Config struct {
	Debug           bool
	EnableLogging   bool
	LoggingDir      string
	SmtpEndpoint    interfaces.SmtpEndpoint    `yaml:"endpoint"`
	SmtpCredentials interfaces.SmtpCredentials `yaml:"credentials"`
	SmtpSubscribers map[string][]string        `yaml:"subscribers"`
	EnableSmtp      bool                       `yaml:"enable-smtp"`
}

type Thread struct {
	Interrupt chan<- common.InterruptEvent // Upon completion or failure an interrupt can be raised

	C3 <-chan common.ThreadRequest  // Messenger is receiving threads form the Database
	C4 chan<- common.ThreadResponse // Messenger is sending responses to the Database

	C17 <-chan common.ThreadRequest // Messenger is receiving requests from the Provisioner

	C22 <-chan common.ThreadRequest  // Messenger is receiving requests from the HTTP Client
	C23 chan<- common.ThreadResponse // Messenger is sending responses to the HTTP Client

	config *Config
	logger *logging.Logger

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, channels ...interface{}) (*Thread, error) {
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
	thread.C3, ok = (channels[1]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 1")
	}
	thread.C4, ok = (channels[2]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 2")
	}
	thread.C17, ok = (channels[3]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	thread.C22, ok = (channels[4]).(chan common.ThreadRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 4")
	}
	thread.C23, ok = (channels[5]).(chan common.ThreadResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 5")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	thread.logger = logger
	thread.logger.SetColour(logging.Blue)

	return thread, nil
}
