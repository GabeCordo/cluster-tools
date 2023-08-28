package http_processor

import (
	"fmt"
	"github.com/GabeCordo/mango/core/threads/common"
	"net/http"
)

func (thread *Thread) Setup() {

	thread.accepting = true

	mux := http.NewServeMux()

	mux.HandleFunc("/processor", func(w http.ResponseWriter, r *http.Request) {
		thread.processorCallback(w, r)
	})

	mux.HandleFunc("/module", func(w http.ResponseWriter, r *http.Request) {
		thread.moduleCallback(w, r)
	})

	mux.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		thread.cacheCallback(w, r)
	})

	mux.HandleFunc("/supervisor", func(w http.ResponseWriter, r *http.Request) {
		thread.supervisorCallback(w, r)
	})

	mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		thread.logCallback(w, r)
	})

	/* the debug endpoint is only enabled when debug is set to true */
	if thread.config.Debug {
		mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
			thread.debugCallback(w, r)
		})
	}

	thread.mux = mux
}

func (thread *Thread) Start() {

	// HTTP API SERVER

	go func(thread *Thread) {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d",
			thread.config.Net.Host, thread.config.Net.Port), thread.mux)
		if err != nil {
			thread.Interrupt <- common.Panic
		}
	}(thread)

	// RESPONSE THREADS

	go func() {
		for response := range thread.C8 {
			thread.ProcessorResponseTable.Write(response.Nonce, response)
		}
	}()

	go func() {
		for response := range thread.C10 {
			thread.CacheResponseTable.Write(response.Nonce, response)
		}
	}()

}

func (thread *Thread) Teardown() {
	thread.accepting = false
}
