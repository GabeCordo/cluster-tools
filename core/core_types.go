package core

import (
	"ETLFramework/logger"
	"ETLFramework/net"
)

type InterruptEvent int8

const (
	Shutdown InterruptEvent = 0
	Panic                   = 1
)

type Thread interface {
	Setup()
	Start()
	Teardown()
}

type Config struct {
	Name    string        `json:"name"`
	Version float64       `json:"version"`
	Debug   bool          `json:"debug"`
	Logging logger.Logger `json:"logging"`
	Net     net.Address   `json:"net"`
	Auth    net.Auth      `json:"auth"`
	Path    string
}

type Core struct {
	httpThread       *Http
	supervisorThread *Supervisor
	messengerThread  *Messenger
	databaseThread   *Database

	c1        chan DatabaseRequest
	c2        chan DatabaseResponse
	c3        chan MessengerRequest
	c4        chan MessengerResponse
	c5        chan SupervisorRequest
	c6        chan SupervisorResponse
	c7        chan DatabaseRequest
	c8        chan DatabaseResponse
	interrupt chan InterruptEvent
}
