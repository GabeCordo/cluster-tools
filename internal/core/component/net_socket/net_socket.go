package net_socket

import (
	"crypto/tls"
	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"net"
)

type NetworkSocket interface {
	Setup(...tls.Certificate) error
	Start(string, int,
		func(uint64, string) error, func(uint64, string) error, func(uint64, *processor2.ModuleConfig), func(*run.Run))
	GetConnection(uint64) (net.Conn, error)
}
