package http

import (
	"context"
	"github.com/Sentmint/cluster-tools/internal/processor/threads"
	"net/http"
	"net/http/pprof"
	"time"
)

func (thread *Thread) Setup() {

	mux := http.NewServeMux()

	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		thread.runCallback(w, r)
	})

	mux.HandleFunc("/core", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		thread.gateCallback(w, r)
	})

	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		thread.debugCallback(w, r)
	})

	mux.HandleFunc("/debug/statistics", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		thread.debugStatsCallback(w, r)
	})

	// TODO - explore this more, fucking cool
	if *thread.Config.Debug {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	thread.mux = mux

	thread.server = &http.Server{
		Addr:        thread.Config.Net,
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
			thread.logger.Println("http thread failed to listen and serve")
			thread.Interrupt <- threads.Panic
		}
	}(thread)

	go func() {
		for response := range thread.C2 {
			if !thread.accepting {
				break
			}
			thread.ProvisionerResponseTable.Write(response.Nonce, response)
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
		thread.logger.Println("teardown called; sending panic")
		thread.Interrupt <- threads.Panic
	}
}
