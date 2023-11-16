package http_client

import (
	"context"
	"fmt"
	"github.com/GabeCordo/mango/core/threads/common"
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

	// TODO - explore this more, fucking cool - removed for now
	if thread.config.Debug {
		mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) { thread.debugCallback(w, r) })
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	thread.mux = mux

	thread.server = &http.Server{
		Addr:        fmt.Sprintf("%s:%d", thread.config.Net.Host, thread.config.Net.Port),
		Handler:     thread.mux,
		ReadTimeout: 2 * time.Second,
	}
	thread.server.SetKeepAlivesEnabled(false)
}

func (thread *Thread) Start() {
	thread.wg.Add(1)

	go func(thread *Thread) {
		err := thread.server.ListenAndServe()
		if err != nil {
			thread.Interrupt <- common.Panic
		}
	}(thread)

	// LISTEN FOR RESPONSES

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

func (thread *Thread) Receive(module common.Module, nonce uint32, timeout ...float64) (any, bool) {
	startTime := time.Now()
	flag := false

	var response any
	for {
		if (len(timeout) > 0) && (time.Now().Sub(startTime).Minutes() > timeout[0]) {
			break
		}

		if module == common.Processor {
			if value, found := thread.supervisorResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		} else if module == common.Database {
			if value, found := thread.databaseResponses[nonce]; found {
				response = value
				flag = true
				break
			}
		}

		// TODO - remove this code
		time.Sleep(2 * time.Millisecond)
	}

	return response, flag
}

func (thread *Thread) Send(module common.Module, request any) (any, bool) {
	thread.mutex.Lock()
	thread.counter++

	nonce := thread.counter // make a copy of the current counter
	if module == common.Processor {
		req := (request).(common.ProcessorRequest)
		req.Nonce = nonce
		thread.C5 <- req
	} else if module == common.Database {
		req := (request).(common.DatabaseRequest)
		req.Nonce = nonce
		thread.C1 <- req
	}

	thread.mutex.Unlock()
	return thread.Receive(module, nonce, thread.config.Timeout)
}

func (thread *Thread) Teardown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := thread.server.Shutdown(ctx)
	if err != nil {
		thread.Interrupt <- common.Panic
	}
}
