package etl

import (
	"ETLFramework/logger"
	"ETLFramework/net"
)

type Segment int8

const (
	Extract   Segment = 0
	Transform         = 1
	Load              = 2
)

type Config struct {
	Name    string        `json:"name"`
	Version float64       `json:"version"`
	Debug   bool          `json:"debug"`
	Logging logger.Logger `json:"logging"`
	Net     net.Address   `json:"net"`
	Auth    net.NodeAuth  `json:"auth"`
}
