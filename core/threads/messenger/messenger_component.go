package messenger

import (
	"github.com/GabeCordo/etl/core/components/messenger"
	"github.com/GabeCordo/etl/core/threads/common"
)

var instance *messenger.Messenger

func GetMessengerInstance() *messenger.Messenger {
	if instance == nil {
		cfg := common.GetConfigInstance()

		instance = messenger.NewMessenger(
			cfg.Messenger.EnableLogging,
			cfg.Messenger.EnableSmtp,
		)

		if cfg.Messenger.EnableLogging {
			instance.LoggingDirectory(cfg.Messenger.LogFiles.Directory)
		}

		if cfg.Messenger.EnableSmtp {
			instance.SetupSMTP(
				messenger.Endpoint{
					Host: cfg.Messenger.Smtp.Endpoint.Host,
					Port: cfg.Messenger.Smtp.Endpoint.Port,
				},
				messenger.Credentials{
					Email:    cfg.Messenger.Smtp.Credentials.Email,
					Password: cfg.Messenger.Smtp.Credentials.Password,
				},
			).SetupReceivers(
				cfg.Messenger.Smtp.Subscribers,
			)
		}
	}
	return instance
}
