package messenger

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/message"
	"github.com/GabeCordo/Flock/internal/core/message/email"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/toolchain/logging"
	"sync"
)

type Config struct {
	Debug           bool
	EnableLogging   bool
	LoggingDir      string
	SmtpEndpoint    email.SmtpEndpoint    `yaml:"endpoint"`
	SmtpCredentials email.SmtpCredentials `yaml:"credentials"`
	SmtpSubscribers map[string][]string   `yaml:"subscribers"`
	EnableSmtp      bool                  `yaml:"enable-smtp"`
}

type Thread struct {
	Interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

	C3 <-chan thread.Request  // Messenger is receiving thread form the Database
	C4 chan<- thread.Response // Messenger is sending responses to the Database

	C17 <-chan thread.Request // Messenger is receiving requests from the provisionerThread

	C22 <-chan thread.Request  // Messenger is receiving requests from the HTTP Client
	C23 chan<- thread.Response // Messenger is sending responses to the HTTP Client

	config *Config
	logger *logging.Logger

	messenger message.Messenger

	accepting bool
	wg        sync.WaitGroup
}

func New(cfg *Config, logger *logging.Logger, messenger message.Messenger, channels ...interface{}) (*Thread, error) {
	th := new(Thread)
	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	th.config = cfg

	th.Interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	th.C3, ok = (channels[1]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 1")
	}
	th.C4, ok = (channels[2]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 2")
	}
	th.C17, ok = (channels[3]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	th.C22, ok = (channels[4]).(chan thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 4")
	}
	th.C23, ok = (channels[5]).(chan thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 5")
	}

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	th.logger = logger
	th.logger.SetColour(logging.Blue)

	th.messenger = messenger

	return th, nil
}
