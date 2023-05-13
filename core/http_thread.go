package core

import (
	"context"
	"net/http"
	"time"
)

//var (
//	NodeInstance   *rpc.Node
//	AuthInstance   *fack.Auth
//	LoggerInstance *logger.Logger
//	nodeLock       = &sync.Mutex{}
//	authLock       = &sync.Mutex{}
//	loggerLock     = &sync.Mutex{}
//)

func (httpThread *HttpThread) Setup() {

	mux := http.NewServeMux()

	mux.HandleFunc("/cluster", func(w http.ResponseWriter, r *http.Request) {
		httpThread.clusterCallback(w, r)
	})

	mux.HandleFunc("/supervisor", func(w http.ResponseWriter, r *http.Request) {
		httpThread.supervisorCallback(w, r)
	})

	mux.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {
		httpThread.statisticCallback(w, r)
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
			httpThread.provisionerResponseTable.Write(supervisorResponse.Nonce, supervisorResponse)
		}
	}()

	go func() {
		for databaseResponse := range httpThread.C2 {
			if !httpThread.accepting {
				break
			}
			httpThread.databaseResponseTable.Write(databaseResponse.Nonce, databaseResponse)
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
