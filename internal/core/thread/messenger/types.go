package messenger

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/component/message"
	"github.com/GabeCordo/Flock/internal/core/component/message/email"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/Flock/internal/shared/logging"
	"github.com/GabeCordo/Flock/internal/shared/terminal"
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
	config   *Config
	channels struct {
		interrupt chan<- thread.InterruptEvent // Upon completion or failure an interrupt can be raised

		c3 <-chan *thread.Request  // Messenger is receiving thread form the Database
		c4 chan<- *thread.Response // Messenger is sending responses to the Database

		c17 <-chan *thread.Request // Messenger is receiving requests from the provisionerThread

		c22 <-chan *thread.Request  // Messenger is receiving requests from the HTTP Client
		c23 chan<- *thread.Response // Messenger is sending responses to the HTTP Client

		close chan thread.InterruptEvent
	}
	logger    logging.Logger
	messenger message.Messenger
}

func New(cfg *Config, logger logging.Logger, messenger message.Messenger, channels ...interface{}) (*Thread, error) {
	th := new(Thread)
	var ok bool

	if cfg == nil {
		return nil, errors.New("expected no nil *pipeline type")
	}
	th.config = cfg

	th.channels.interrupt, ok = (channels[0]).(chan thread.InterruptEvent)
	if !ok {
		return nil, errors.New("expected type 'chan InterruptEvent' in index 0")
	}
	th.channels.c3, ok = (channels[1]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 1")
	}
	th.channels.c4, ok = (channels[2]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 2")
	}
	th.channels.c17, ok = (channels[3]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 3")
	}
	th.channels.c22, ok = (channels[4]).(chan *thread.Request)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerRequest' in index 4")
	}
	th.channels.c23, ok = (channels[5]).(chan *thread.Response)
	if !ok {
		return nil, errors.New("expected type 'chan MessengerResponse' in index 5")
	}
	th.channels.close = make(chan thread.InterruptEvent)

	if logger == nil {
		return nil, errors.New("expected non nil *utils.logger type")
	}
	th.logger = logger
	th.logger.SetColour(terminal.Blue)

	th.messenger = messenger

	return th, nil
}
