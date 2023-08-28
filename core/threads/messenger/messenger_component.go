package messenger

import (
	"github.com/GabeCordo/mango/core/components/messenger"
)

var instance *messenger.Messenger

func GetMessengerInstance(cfg *Config) *messenger.Messenger {
	if instance == nil {

		instance = messenger.NewMessenger(
			cfg.EnableLogging,
			cfg.EnableSmtp,
		)

		if cfg.EnableLogging {
			instance.LoggingDirectory(cfg.LoggingDir)
		}

		if cfg.EnableSmtp {
			instance.SetupSMTP(
				messenger.Endpoint{
					Host: cfg.SmtpEndpoint.Host,
					Port: cfg.SmtpEndpoint.Port,
				},
				messenger.Credentials{
					Email:    cfg.SmtpCredentials.Email,
					Password: cfg.SmtpCredentials.Password,
				},
			).SetupReceivers(
				cfg.SmtpSubscribers,
			)
		}
	}
	return instance
}
