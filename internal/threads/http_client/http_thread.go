package http_client

import (
	"context"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
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

	mux.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		thread.jobCallback(w, r)
	})

	mux.HandleFunc("/job/queue", func(w http.ResponseWriter, r *http.Request) {
		thread.jobQueueCallback(w, r)
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

	go func() {
		for schedulerResponse := range thread.C21 {
			if !thread.accepting {
				break
			}
			thread.SchedulerResponseTable.Write(schedulerResponse.Nonce, schedulerResponse)
		}
	}()

	go func() {
		for messengerResponse := range thread.C23 {
			if !thread.accepting {
				break
			}
			thread.MessengerResponseTable.Write(messengerResponse.Nonce, messengerResponse)
		}
	}()

	go func() {
		for cacheResponse := range thread.C25 {
			if !thread.accepting {
				break
			}
			thread.SchedulerResponseTable.Write(cacheResponse.Nonce, cacheResponse)
		}
	}()

	thread.wg.Wait()
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
