package core

import (
	"ETLFramework/logger"
	"ETLFramework/net"
	"log"
	"sync"
)

var (
	NodeInstance      *net.Node
	AuthInstance      *net.Auth
	LoggerInstance    *logger.Logger
	commonLoggerPaths = [...]string{"/logs"}
	nodeLock          = &sync.Mutex{}
	authLock          = &sync.Mutex{}
	loggerLock        = &sync.Mutex{}
)

func GetNodeInstance() *net.Node {
	nodeLock.Lock()
	defer nodeLock.Unlock()

	if NodeInstance == nil {
		config := GetConfigInstance()
		NodeInstance = net.NewNode(config.Net, config.Debug, GetAuthInstance()) // TODO - re-add the logger at a later date
		NodeInstance.Name = config.Name
	}

	return NodeInstance
}

func GetAuthInstance() *net.Auth {
	authLock.Lock()
	defer authLock.Unlock()

	if AuthInstance == nil {
		AuthInstance = &GetConfigInstance().Auth

		// the config may not define a map of trusted endpoints leaving the
		// Trusted field as a nil value that cannot be used
		if AuthInstance.Trusted == nil {
			AuthInstance.Trusted = make(map[string]*net.Endpoint)
		}

		// ECDSA public keys are stored as an uint64 representation of bytes
		// to ease the process of copying + storing keys - convert to the ECDSA structure
		for trusted, endpoint := range AuthInstance.Trusted {
			_, ok := endpoint.GetPublicKey() // populates the PublicKey structure using the uint64 bytes
			if !ok {
				log.Println("failed to generate ECDSA key for trusted " + trusted)
			}
		}
	}

	return AuthInstance
}

func GetLoggerInstance() *logger.Logger {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	if LoggerInstance == nil {
		LoggerInstance = &GetConfigInstance().Logging

		// we may not contain a JSON mapping of the logging queue, meaning a nil
		// value will hold its place that can raise an error
		if LoggerInstance.LogQueue == nil {
			LoggerInstance.LogQueue = make(chan string)
		}

		if len(LoggerInstance.Folder) == 0 {
			LoggerInstance.Folder = "/logs" // TODO - implement a platform specific way to create logs
		}
	}

	return LoggerInstance
}

func (http Http) Setup() {
	//logg = frontend.GetLoggerInstance()
	node := GetNodeInstance()
	// core_callbacks functions
	node.Function("/clusters", http.ClustersFunction, []string{"GET"}, false)
	node.Function("/statistics", http.StatisticsFunction, []string{"GET"}, true)
	node.Function("/debug", http.DebugFunction, []string{"GET", "POST", "DELETE"}, true)
}

func (http Http) Start() {
	//go logg.LoggerEventLoop() // TODO - consider renaming this
	//logg.Alert("core", "logging started")
	GetNodeInstance().Start()
}

func (http Http) Teardown() {
	GetNodeInstance().Shutdown()
}
