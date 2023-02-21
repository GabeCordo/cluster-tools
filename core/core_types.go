package core

import (
	"github.com/GabeCordo/etl/components/messenger"
	"github.com/GabeCordo/fack"
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
)

type Thread interface {
	Setup()
	Start()
	Teardown()
}

type Config struct {
	Name              string   `json:"name"`
	Version           float64  `json:"version"`
	Debug             bool     `json:"debug"`
	HardTerminateTime int      `json:"hard-terminate-time"`
	AutoMount         []string `json:"auto-mount"`
	Cache             struct {
		Expiry  float64 `json:"expire-in"`
		MaxSize uint32  `json:"max-size"`
	} `json:"cache"`
	Messenger struct {
		LogFiles struct {
			Directory string `json:"directory"`
		} `json:"logging,omitempty"`
		EnableLogging bool `json:"enable-logging"`
		Smtp          struct {
			Endpoint    messenger.Endpoint    `json:"endpoint"`
			Credentials messenger.Credentials `json:"credentials"`
			Subscribers map[string][]string   `json:"subscribers"`
		} `json:"smtp,omitempty"`
		EnableSmtp bool `json:"enable-smtp"`
	} `json:"messenger"`
	Net  fack.Address `json:"net"`
	Auth fack.Auth    `json:"auth"`
	Path string
}

func (c *Config) Safe() *Config {
	if c.AutoMount == nil {
		c.AutoMount = make([]string, 0)
	}

	return c
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
}
