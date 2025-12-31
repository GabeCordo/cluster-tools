package rest

import (
	"context"
	"fmt"
	"github.com/FortifiedCode/flock/internal/flags"
	"github.com/FortifiedCode/flock/internal/shared/terminal"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
	"net/http"
	"net/http/pprof"
	"time"
)

func (t *Thread) Setup() {

	t.logger.SetColour(terminal.Green)

	mux := http.NewServeMux()

	f := func(b func(w http.ResponseWriter, req *http.Request), w http.ResponseWriter, r *http.Request) {
		t.logger.Printf("[%s] %s?%s\n", r.Method, r.URL.Path, r.URL.RawQuery)
		b(w, r)
	}

	mux.HandleFunc("/processor", func(w http.ResponseWriter, r *http.Request) {
		f(t.processorCallback, w, r)
	})

	mux.HandleFunc("/module", func(w http.ResponseWriter, r *http.Request) {
		f(t.moduleCallback, w, r)
	})

	mux.HandleFunc("/function", func(w http.ResponseWriter, r *http.Request) {
		f(t.functionCallback, w, r)
	})

	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		f(t.runCallback, w, r)
	})

	mux.HandleFunc("/run/count", func(w http.ResponseWriter, r *http.Request) {
		f(t.runCountCallback, w, r)
	})

	mux.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {
		f(t.statisticsCallback, w, r)
	})

	mux.HandleFunc("/statistic/", func(w http.ResponseWriter, r *http.Request) {
		f(t.statisticCallback, w, r)
	})

	mux.HandleFunc("/namespaces/", func(w http.ResponseWriter, r *http.Request) {
		f(t.namespaceCallback, w, r)
	})

	mux.HandleFunc("/pipeline", func(w http.ResponseWriter, r *http.Request) {
		f(t.pipelineCallback, w, r)
	})

	mux.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		f(t.jobCallback, w, r)
	})

	mux.HandleFunc("/job/queue", func(w http.ResponseWriter, r *http.Request) {
		f(t.jobQueueCallback, w, r)
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

func (t *Thread) HandleRequest(request *thread.Request) {
	// note: this isn't needed at this time but keep the plain definition to satisfy the interface
	panic("implement me")
}

func (t *Thread) Start() {

	go func(t *Thread) {
		err := t.server.ListenAndServe()
		if err != nil {
			t.logger.Alertf("The REST API is unable to bind to the port %d", t.config.Net.Port)
			t.logger.Alertf("You need to check what service is using port %d", t.config.Net.Port)
			t.channels.interrupt <- thread.Panic
		}
	}(t)

	var iRsp *thread.Response

	for {
		select {
		case iRsp = <-t.channels.c6:
			{
				if flags.DEBUG {
					t.logger.Printf("Received %s", iRsp.ToString())
				}
				t.ProcessorResponseTable.Write(iRsp.Nonce, iRsp)
			}
		case iRsp = <-t.channels.c2:
			{
				if flags.DEBUG {
					t.logger.Printf("Received %s", iRsp.ToString())
				}
				t.DatabaseResponseTable.Write(iRsp.Nonce, iRsp)
			}
		case iRsp = <-t.channels.c21:
			{
				if flags.DEBUG {
					t.logger.Printf("Received %s", iRsp.ToString())
				}
				t.SchedulerResponseTable.Write(iRsp.Nonce, iRsp)
			}
		case iRsp = <-t.channels.c23:
			{
				if flags.DEBUG {
					t.logger.Printf("Received %s", iRsp.ToString())
				}
				t.MessengerResponseTable.Write(iRsp.Nonce, iRsp)
			}
		case <-t.channels.close:
			{
				// shutting down the rest thread
				break
			}
		}
	}
}

func (t *Thread) Teardown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := t.server.Shutdown(ctx)
	if err != nil {
		t.channels.interrupt <- thread.Panic
	}

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
