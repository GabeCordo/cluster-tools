package core

import (
	"github.com/GabeCordo/etl/components/messenger"
	"github.com/GabeCordo/etl/components/utils"
)

type InterruptEvent uint8

const (
	Shutdown InterruptEvent = 0
	Panic                   = 1
)

type Module uint8

const (
	Http        Module = 0
	Database           = 1
	Provisioner        = 2
	Messenger          = 3
	Cache              = 4
)

type Thread interface {
	Setup()
	Start()
	Teardown()
}

type Config struct {
	Name               string  `yaml:"name"`
	Version            float64 `yaml:"version"`
	Debug              bool    `yaml:"debug"`
	HardTerminateTime  int     `yaml:"hard-terminate-time"`
	MaxWaitForResponse float64 `yaml:"max-wait-for-response"`
	Cache              struct {
		Expiry  float64 `yaml:"expire-in"`
		MaxSize uint32  `yaml:"max-size"`
	} `yaml:"cache"`
	Messenger struct {
		LogFiles struct {
			Directory string `yaml:"directory"`
		} `yaml:"logging,omitempty"`
		EnableLogging bool `yaml:"enable-logging"`
		Smtp          struct {
			Endpoint    messenger.Endpoint    `yaml:"endpoint"`
			Credentials messenger.Credentials `yaml:"credentials"`
			Subscribers map[string][]string   `yaml:"subscribers"`
		} `json:"smtp,omitempty"`
		EnableSmtp bool `yaml:"enable-smtp"`
	} `json:"messenger"`
	Net struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"net"`
	Path string
}

type Core struct {
	HttpThread        *HttpThread
	ProvisionerThread *ProvisionerThread
	MessengerThread   *MessengerThread
	DatabaseThread    *DatabaseThread
	CacheThread       *CacheThread

	C1        chan DatabaseRequest
	C2        chan DatabaseResponse
	C3        chan MessengerRequest
	C4        chan MessengerResponse
	C5        chan ProvisionerRequest
	C6        chan ProvisionerResponse
	C7        chan DatabaseRequest
	C8        chan DatabaseResponse
	C9        chan CacheRequest
	C10       chan CacheResponse
	C11       chan MessengerRequest
	interrupt chan InterruptEvent

	logger *utils.Logger
}
