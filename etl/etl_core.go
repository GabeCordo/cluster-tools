package etl

import (
	"ETLFramework/logger"
	"ETLFramework/net"
	"sync"
)

const (
	ConfigPath = "../configs/config.etl.json"
)

var configLock = &sync.Mutex{}
var nodeLock = &sync.Mutex{}
var authLock = &sync.Mutex{}
var loggerLock = &sync.Mutex{}

var (
	ConfigInstance *Config
	NodeInstance   *net.Node
	AuthInstance   *net.Auth
	LoggerInstance *logger.Logger
)

func GetConfigInstance() *Config {
	configLock.Lock()
	defer configLock.Unlock()

	if ConfigInstance == nil {
		ConfigInstance = new(Config)
		err := JSONToETLConfig(ConfigInstance, ConfigPath)
		if err != nil {
			panic(err)
		}
	}

	return ConfigInstance
}

func GetNodeInstance() *net.Node {
	nodeLock.Lock()
	defer nodeLock.Unlock()

	if NodeInstance == nil {
		config := GetConfigInstance()
		NodeInstance = net.NewNode(config.Name, config.Net, config.Debug, GetAuthInstance(), GetLoggerInstance())
	}

	return NodeInstance
}

func GetAuthInstance() *net.Auth {
	authLock.Lock()
	defer authLock.Unlock()

	if AuthInstance == nil {
		AuthInstance = &GetConfigInstance().Auth
	}

	return AuthInstance
}

func GetLoggerInstance() *logger.Logger {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	if LoggerInstance == nil {
		LoggerInstance = &GetConfigInstance().Logging
	}

	return LoggerInstance
}
