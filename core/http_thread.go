package core

import (
	"ETLFramework/logger"
	"ETLFramework/net"
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

func (http Http) ClustersFunction(request *net.Request, response *net.Response) {
	supervisorRequest := SupervisorRequest{Provision, request.Function, request.Param}
	http.C5 <- supervisorRequest

	response.AddStatus(200, "good")
}

func (http Http) StatisticsFunction(request *net.Request, response *net.Response) {
	// TODO - not implemented
}

func (http Http) DebugFunction(request *net.Request, response *net.Response) {
	if request.Function == "shutdown" {
		http.Interrupt <- Shutdown
	}
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
