package core

import (
	"ETLFramework/logger"
	"ETLFramework/net"
	"log"
	"sync"
	"time"
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

func (http HttpThread) Setup() {
	//logg = frontend.GetLoggerInstance()
	node := GetNodeInstance()
	// core_callbacks functions
	node.Function("/clusters", http.ClustersFunction, []string{"GET"}, false)
	node.Function("/statistics", http.StatisticsFunction, []string{"GET"}, true)
	node.Function("/data", http.DataFunction, []string{"GET"}, true)
	node.Function("/debug", http.DebugFunction, []string{"GET", "POST", "DELETE"}, true)
}

func (http *HttpThread) Start() {
	http.wg.Add(1)

	go GetNodeInstance().Start()

	go func() {
		for supervisorResponse := range http.C6 {
			if !http.accepting {
				break
			}
			http.supervisorResponses[supervisorResponse.Nonce] = supervisorResponse
		}
	}()

	go func() {
		for databaseResponse := range http.C2 {
			if !http.accepting {
				break
			}
			http.databaseResponses[databaseResponse.Nonce] = databaseResponse
		}
	}()

	http.wg.Wait()
}

func (http *HttpThread) Receive(module Module, nonce uint32, timeout ...float64) (any, bool) {
	startTime := time.Now()
	flag := false

	var response any
	for {
		if (len(timeout) > 0) && (time.Now().Sub(startTime).Minutes() > timeout[0]) {
			break
		}

		if module == Provisioner {
			if value, found := http.supervisorResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		} else if module == Database {
			if value, found := http.databaseResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		}

		time.Sleep(RefreshTime * time.Millisecond)
	}

	return response, flag
}

func (http *HttpThread) Send(module Module, request any) (any, bool) {
	http.mutex.Lock()
	http.counter++

	nonce := http.counter // make a copy of the current counter
	if module == Provisioner {
		req := (request).(ProvisionerRequest)
		req.Nonce = nonce
		http.C5 <- req
	} else if module == Database {
		req := (request).(DatabaseRequest)
		req.Nonce = nonce
		http.C1 <- req
	}

	http.mutex.Unlock()
	return http.Receive(module, nonce, DefaultTimeout)
}

func (http *HttpThread) Teardown() {
	GetNodeInstance().Shutdown()
}
