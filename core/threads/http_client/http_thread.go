package http_client

import (
	"context"
	"fmt"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"net/http"
	"net/http/pprof"
	"time"
)

func (thread *Thread) Setup() {

	mux := http.NewServeMux()

	mux.HandleFunc("/processor", func(w http.ResponseWriter, r *http.Request) {
		thread.processorCallback(w, r)
	})

	mux.HandleFunc("/module", func(w http.ResponseWriter, r *http.Request) {
		thread.moduleCallback(w, r)
	})

	mux.HandleFunc("/cluster", func(w http.ResponseWriter, r *http.Request) {
		thread.clusterCallback(w, r)
	})

	mux.HandleFunc("/supervisor", func(w http.ResponseWriter, r *http.Request) {
		thread.supervisorCallback(w, r)
	})

	mux.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {
		thread.statisticCallback(w, r)
	})

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		thread.configCallback(w, r)
	})

	// TODO - explore this more, fucking cool
	if common.GetConfigInstance().Debug {
		mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) { thread.debugCallback(w, r) })
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	if common.GetConfigInstance().EnableCors {
		mux.HandleFunc("/cors", thread.corsCallback)
	}

	thread.mux = mux
}

func (thread *Thread) Start() {
	thread.wg.Add(1)

	go func(thread *Thread) {
		net := common.GetConfigInstance().Net.Client
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", net.Host, net.Port), thread.mux)
		if err != nil {
			thread.Interrupt <- threads.Panic
		}
	}(thread)

	go func() {
		for supervisorResponse := range thread.C6 {
			if !thread.accepting {
				break
			}
			thread.ProcessorResponseTable.Write(supervisorResponse.Nonce, supervisorResponse)
		}
	}()

	go func() {
		for databaseResponse := range thread.C2 {
			if !thread.accepting {
				break
			}
			thread.DatabaseResponseTable.Write(databaseResponse.Nonce, databaseResponse)
		}
	}()

	thread.wg.Wait()
}

func (thread *Thread) Receive(module threads.Module, nonce uint32, timeout ...float64) (any, bool) {
	startTime := time.Now()
	flag := false

	var response any
	for {
		if (len(timeout) > 0) && (time.Now().Sub(startTime).Minutes() > timeout[0]) {
			break
		}

		if module == threads.Provisioner {
			if value, found := thread.supervisorResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		} else if module == threads.Database {
			if value, found := thread.databaseResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		}

		time.Sleep(threads.RefreshTime * time.Millisecond)
	}

	return response, flag
}

func (thread *Thread) Send(module threads.Module, request any) (any, bool) {
	thread.mutex.Lock()
	thread.counter++

	nonce := thread.counter // make a copy of the current counter
	if module == threads.Provisioner {
		req := (request).(common.ProcessorRequest)
		req.Nonce = nonce
		thread.C5 <- req
	} else if module == threads.Database {
		req := (request).(threads.DatabaseRequest)
		req.Nonce = nonce
		thread.C1 <- req
	}

	thread.mutex.Unlock()
	return thread.Receive(module, nonce, threads.DefaultTimeout)
}

func (thread *Thread) Teardown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := thread.server.Shutdown(ctx)
	if err != nil {
		thread.Interrupt <- threads.Panic
	}
}
