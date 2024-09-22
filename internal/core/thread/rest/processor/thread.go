package processor

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
	"net/http"
	"time"
)

func (t *Thread) Setup() {

	t.accepting = true

	mux := http.NewServeMux()

	mux.HandleFunc("/processor", func(w http.ResponseWriter, r *http.Request) {
		t.processorCallback(w, r)
		r.Body.Close()
	})

	mux.HandleFunc("/module", func(w http.ResponseWriter, r *http.Request) {
		t.moduleCallback(w, r)
		r.Body.Close()
	})

	mux.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		t.cacheCallback(w, r)
		r.Body.Close()
	})

	mux.HandleFunc("/supervisor", func(w http.ResponseWriter, r *http.Request) {
		t.supervisorCallback(w, r)
		r.Body.Close()
	})

	mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		t.logCallback(w, r)
		r.Body.Close()
	})

	/* the debug endpoint is only enabled when debug is set to true */
	if t.config.Debug {
		mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
			t.debugCallback(w, r)
		})
	}

	t.mux = mux

	t.server = &http.Server{
		Addr:        fmt.Sprintf("%s:%d", t.config.Net.Host, t.config.Net.Port),
		ReadTimeout: 2 * time.Second,
		Handler:     t.mux,
	}
	t.server.SetKeepAlivesEnabled(false)
}

func (t *Thread) Start() {

	// HTTP API SERVER

	go func(t *Thread) {

		err := t.server.ListenAndServe()
		if err != nil {
			t.Interrupt <- thread.Panic
		}
	}(t)

	// RESPONSE THREADS

	go func() {
		for response := range t.C8 {
			t.ProcessorResponseTable.Write(response.Nonce, response)
		}
	}()

	go func() {
		for response := range t.C10 {
			t.CacheResponseTable.Write(response.Nonce, response)
		}
	}()

}

func (t *Thread) Handle(request *thread.Request, response *thread.Response) {
	// note: this isn't needed at this time but keep the plain definition to satisfy the interface
	panic("implement me")
}

func (t *Thread) Teardown() {
	t.accepting = false
}
