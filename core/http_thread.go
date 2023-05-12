package core

import (
	"context"
	"fmt"
	"github.com/GabeCordo/etl/components/logger"
	"github.com/GabeCordo/etl/components/utils"
	"github.com/GabeCordo/fack"
	"github.com/GabeCordo/fack/rpc"
	"net/http"
	"sync"
	"time"
)

var (
	NodeInstance   *rpc.Node
	AuthInstance   *fack.Auth
	LoggerInstance *logger.Logger
	nodeLock       = &sync.Mutex{}
	authLock       = &sync.Mutex{}
	loggerLock     = &sync.Mutex{}
)

var (
	provisionerResponseTable *utils.ResponseTable
	databaseResponseTable    *utils.ResponseTable
)

func GetProvisionerResponseTable() *utils.ResponseTable {
	if provisionerResponseTable == nil {
		provisionerResponseTable = utils.NewResponseTable()
	}
	return provisionerResponseTable
}

func GetDatabaseResponseTable() *utils.ResponseTable {
	if databaseResponseTable == nil {
		databaseResponseTable = utils.NewResponseTable()
	}
	return databaseResponseTable
}

//
//func GetNodeInstance() *rpc.Node {
//	nodeLock.Lock()
//	defer nodeLock.Unlock()
//
//	if NodeInstance == nil {
//		config := GetConfigInstance()
//		NodeInstance = rpc.NewNode(&config.Net, config.Debug, GetAuthInstance()) // TODO - re-add the logger at a later date
//		NodeInstance.Name(config.Name)
//	}
//
//	return NodeInstance
//}
//
//func GetAuthInstance() *fack.Auth {
//	authLock.Lock()
//	defer authLock.Unlock()
//
//	if AuthInstance == nil {
//		AuthInstance = &GetConfigInstance().Auth
//
//		// the config may not define a map of trusted endpoints leaving the
//		// Trusted field as a nil value that cannot be used
//		if AuthInstance.Trusted == nil {
//			AuthInstance.Trusted = make(map[string]*fack.Endpoint)
//		}
//
//		// ECDSA public keys are stored as an uint64 representation of bytes
//		// to ease the process of copying + storing keys - convert to the ECDSA structure
//		for trusted, endpoint := range AuthInstance.Trusted {
//			_, ok := endpoint.GetPublicKey() // populates the PublicKey structure using the uint64 bytes
//			if !ok {
//				log.Println("failed to generate ECDSA key for trusted " + trusted)
//			}
//		}
//	}
//
//	return AuthInstance
//}
//
//func GetLoggerInstance() *logger.Logger {
//	loggerLock.Lock()
//	defer loggerLock.Unlock()
//
//	if LoggerInstance == nil {
//		// TODO - allow the logger to be customized
//		LoggerInstance = logger.NewLogger(ConfigInstance.Path, logger.Verbose, logger.NewInterval(0, 10))
//
//		// we may not contain a JSON mapping of the logging queue, meaning a nil
//		// value will hold its place that can raise an error
//		if LoggerInstance.LogQueue == nil {
//			LoggerInstance.LogQueue = make(chan string)
//		}
//
//		if len(LoggerInstance.Folder) == 0 {
//			LoggerInstance.Folder = "/logs" // TODO - implement a platform specific way to create logs
//		}
//	}
//
//	return LoggerInstance
//}

const keyServerAddr = "serverAddr"

func (httpThread *HttpThread) Setup() {

	mux := http.NewServeMux()

	mux.HandleFunc("/cluster", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("test")
		httpThread.clusterCallback(w, r)
	})

	mux.HandleFunc("/supervisor", func(w http.ResponseWriter, r *http.Request) {
		httpThread.supervisorCallback(w, r)
	})

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		httpThread.configCallback(w, r)
	})

	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		httpThread.debugCallback(w, r)
	})

	httpThread.mux = mux
}

func (httpThread *HttpThread) Start() {
	httpThread.wg.Add(1)

	go func(thread *HttpThread) {
		err := http.ListenAndServe(GetConfigInstance().Net.ToString(), httpThread.mux)
		if err != nil {
			thread.Interrupt <- Panic
		}
	}(httpThread)

	go func() {
		for supervisorResponse := range httpThread.C6 {
			if !httpThread.accepting {
				break
			}
			GetProvisionerResponseTable().Write(supervisorResponse.Nonce, supervisorResponse)
		}
	}()

	go func() {
		for databaseResponse := range httpThread.C2 {
			if !httpThread.accepting {
				break
			}
			fmt.Println(databaseResponse)
			GetDatabaseResponseTable().Write(databaseResponse.Nonce, databaseResponse)
		}
	}()

	httpThread.wg.Wait()
}

func (httpThread *HttpThread) Receive(module Module, nonce uint32, timeout ...float64) (any, bool) {
	startTime := time.Now()
	flag := false

	var response any
	for {
		if (len(timeout) > 0) && (time.Now().Sub(startTime).Minutes() > timeout[0]) {
			break
		}

		if module == Provisioner {
			if value, found := httpThread.supervisorResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		} else if module == Database {
			if value, found := httpThread.databaseResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		}

		time.Sleep(RefreshTime * time.Millisecond)
	}

	return response, flag
}

func (httpThread *HttpThread) Send(module Module, request any) (any, bool) {
	httpThread.mutex.Lock()
	httpThread.counter++

	nonce := httpThread.counter // make a copy of the current counter
	if module == Provisioner {
		req := (request).(ProvisionerRequest)
		req.Nonce = nonce
		httpThread.C5 <- req
	} else if module == Database {
		req := (request).(DatabaseRequest)
		req.Nonce = nonce
		httpThread.C1 <- req
	}

	httpThread.mutex.Unlock()
	return httpThread.Receive(module, nonce, DefaultTimeout)
}

func (httpThread *HttpThread) Teardown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := httpThread.server.Shutdown(ctx)
	if err != nil {
		httpThread.Interrupt <- Panic
	}
}
