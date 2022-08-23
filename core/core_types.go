package core

import (
	"ETLFramework/logger"
	"ETLFramework/net"
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
	Name              string        `json:"name"`
	Version           float64       `json:"version"`
	Debug             bool          `json:"debug"`
	HardTerminateTime int           `json:"hard-terminate-time"`
	AutoMount         []string      `json:"auto-mount"`
	Logging           logger.Logger `json:"logging"`
	Net               net.Address   `json:"net"`
	Auth              net.Auth      `json:"auth"`
	Path              string
}

type Core struct {
	HttpThread        *HttpThread
	ProvisionerThread *ProvisionerThread
	MessengerThread   *MessengerThread
	DatabaseThread    *DatabaseThread

	c1        chan DatabaseRequest
	c2        chan DatabaseResponse
	c3        chan MessengerRequest
	c4        chan MessengerResponse
	c5        chan ProvisionerRequest
	c6        chan ProvisionerResponse
	c7        chan DatabaseRequest
	c8        chan DatabaseResponse
	interrupt chan InterruptEvent
}
