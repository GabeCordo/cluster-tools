package http

import (
	"context"
	"fmt"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/framework/core/common"
	"net/http"
	"net/http/pprof"
	"time"
)

func (httpThread *Thread) Setup() {

	mux := http.NewServeMux()

	mux.HandleFunc("/module", func(w http.ResponseWriter, r *http.Request) {
		httpThread.moduleCallback(w, r)
	})

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

	// TODO - explore this more, fucking cool
	if common.GetConfigInstance().Debug {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	httpThread.mux = mux
}

func (httpThread *Thread) Start() {
	httpThread.wg.Add(1)

	go func(thread *Thread) {
		net := common.GetConfigInstance().Net
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", net.Host, net.Port), httpThread.mux)
		if err != nil {
			fmt.Println("could not start http server")
			thread.Interrupt <- threads.Panic
		}
	}(httpThread)

	for httpThread.accepting {

		select {
		case response := <-httpThread.C6:
			httpThread.ProvisionerResponseTable.Write(response.Nonce, response)
		case response := <-httpThread.C2:
			httpThread.DatabaseResponseTable.Write(response.Nonce, response)
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	httpThread.wg.Wait()
}

func (httpThread *Thread) Receive(module threads.Module, nonce uint32, timeout ...float64) (any, bool) {
	startTime := time.Now()
	flag := false

	var response any
	for {
		if (len(timeout) > 0) && (time.Now().Sub(startTime).Minutes() > timeout[0]) {
			break
		}

		if module == threads.Provisioner {
			if value, found := httpThread.supervisorResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		} else if module == threads.Database {
			if value, found := httpThread.databaseResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		}

		time.Sleep(threads.RefreshTime * time.Millisecond)
	}

	return response, flag
}

func (httpThread *Thread) Send(module threads.Module, request any) (any, bool) {
	httpThread.mutex.Lock()
	httpThread.counter++

	nonce := httpThread.counter // make a copy of the current counter
	if module == threads.Provisioner {
		req := (request).(threads.ProvisionerRequest)
		req.Nonce = nonce
		httpThread.C5 <- req
	} else if module == threads.Database {
		req := (request).(threads.DatabaseRequest)
		req.Nonce = nonce
		httpThread.C1 <- req
	}

	httpThread.mutex.Unlock()
	return httpThread.Receive(module, nonce, threads.DefaultTimeout)
}

func (httpThread *Thread) Teardown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := httpThread.server.Shutdown(ctx)
	if err != nil {
		httpThread.Interrupt <- threads.Panic
	}
}
