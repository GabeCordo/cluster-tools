package messenger

import (
	"errors"
	"github.com/GabeCordo/mango/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"sync"
)

type Config struct {
	Debug         bool
	EnableLogging bool
	LoggingDir    string
	SmtpEndpoint  struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"endpoint"`
	SmtpCredentials struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"credentials"`
	SmtpSubscribers map[string][]string `yaml:"subscribers"`
	EnableSmtp      bool                `yaml:"enable-smtp"`
}

type Thread struct {
	Interrupt chan<- common.InterruptEvent // Upon completion or failure an interrupt can be raised

	C3  <-chan common.MessengerRequest  // Messenger is receiving threads form the Database
	C4  chan<- common.MessengerResponse // Messenger is sending responses to the Database
	C17 <-chan common.MessengerRequest  // Messenger is receiving requests from the Provisioner

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
	thread.C3, ok = (channels[1]).(chan common.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 1")
	}
	thread.C4, ok = (channels[2]).(chan common.MessengerResponse)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 2")
	}
	thread.C17, ok = (channels[3]).(chan common.MessengerRequest)
	if !ok {
		return nil, errors.New("expected type 'chan MesengerRequest' in index 3")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.Logger type")
	}
	thread.logger = logger
	thread.logger.SetColour(logging.Blue)

	return thread, nil
}
