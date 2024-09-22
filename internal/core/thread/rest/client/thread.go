package client

import (
	"context"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
	"net/http"
	"net/http/pprof"
	"time"
)

func (t *Thread) Setup() {

	mux := http.NewServeMux()

	mux.HandleFunc("/processor", func(w http.ResponseWriter, r *http.Request) {
		t.processorCallback(w, r)
	})

	mux.HandleFunc("/module", func(w http.ResponseWriter, r *http.Request) {
		t.moduleCallback(w, r)
	})

	mux.HandleFunc("/function", func(w http.ResponseWriter, r *http.Request) {
		t.functionCallback(w, r)
	})

	mux.HandleFunc("/supervisor", func(w http.ResponseWriter, r *http.Request) {
		t.supervisorCallback(w, r)
	})

	mux.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {
		t.statisticCallback(w, r)
	})

	mux.HandleFunc("/pipeline", func(w http.ResponseWriter, r *http.Request) {
		t.pipelineCallback(w, r)
	})

	mux.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		t.jobCallback(w, r)
	})

	mux.HandleFunc("/job/queue", func(w http.ResponseWriter, r *http.Request) {
		t.jobQueueCallback(w, r)
	})

	// TODO - explore this more, fucking cool - removed for now
	if t.config.Debug {
		mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) { t.debugCallback(w, r) })
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	t.mux = mux

	t.server = &http.Server{
		Addr:        fmt.Sprintf("%s:%d", t.config.Net.Host, t.config.Net.Port),
		Handler:     t.mux,
		ReadTimeout: 2 * time.Second,
	}
	t.server.SetKeepAlivesEnabled(false)
}

func (t *Thread) Handle(request *thread.Request, response *thread.Response) {
	// note: this isn't needed at this time but keep the plain definition to satisfy the interface
	panic("implement me")
}

func (t *Thread) Start() {
	t.wg.Add(1)

	go func(t *Thread) {
		err := t.server.ListenAndServe()
		if err != nil {
			t.Interrupt <- thread.Panic
		}
	}(t)

	// LISTEN FOR RESPONSES

	go func() {
		for supervisorResponse := range t.C6 {
			if !t.accepting {
				break
			}
			t.ProcessorResponseTable.Write(supervisorResponse.Nonce, supervisorResponse)
		}
	}()

	go func() {
		for databaseResponse := range t.C2 {
			if !t.accepting {
				break
			}
			t.DatabaseResponseTable.Write(databaseResponse.Nonce, databaseResponse)
		}
	}()

	go func() {
		for schedulerResponse := range t.C21 {
			if !t.accepting {
				break
			}
			t.SchedulerResponseTable.Write(schedulerResponse.Nonce, schedulerResponse)
		}
	}()

	go func() {
		for messengerResponse := range t.C23 {
			if !t.accepting {
				break
			}
			t.MessengerResponseTable.Write(messengerResponse.Nonce, messengerResponse)
		}
	}()

	go func() {
		for cacheResponse := range t.C25 {
			if !t.accepting {
				break
			}
			t.SchedulerResponseTable.Write(cacheResponse.Nonce, cacheResponse)
		}
	}()

	t.wg.Wait()
}

func (t *Thread) Teardown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := t.server.Shutdown(ctx)
	if err != nil {
		t.Interrupt <- thread.Panic
	}
}
