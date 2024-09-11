package messenger

import (
	"github.com/GabeCordo/cluster-tools/internal/core/components/directory"
	"github.com/GabeCordo/cluster-tools/internal/core/components/messenger"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
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
				interfaces.SmtpEndpoint{
					Host: cfg.SmtpEndpoint.Host,
					Port: cfg.SmtpEndpoint.Port,
				},
				interfaces.SmtpCredentials{
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

var contactDirectoryInstance *directory.Directory

func GetContactDirectory() *directory.Directory {

	if contactDirectoryInstance == nil {
		contactDirectoryInstance = directory.NewDirectory()
	}

	return contactDirectoryInstance
}
