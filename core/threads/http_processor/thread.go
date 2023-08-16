package http_processor

import (
	"fmt"
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl/core/threads/common"
	"net/http"
)

func (thread *Thread) Setup() {

	mux := http.NewServeMux()

	mux.HandleFunc("/processor", func(w http.ResponseWriter, r *http.Request) {
		thread.processorCallback(w, r)
	})

	mux.HandleFunc("/module", func(w http.ResponseWriter, r *http.Request) {
		thread.moduleCallback(w, r)
	})

	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		thread.debugCallback(w, r)
	})

	thread.mux = mux
}

func (thread *Thread) Start() {

	go func(thread *Thread) {
		net := common.GetConfigInstance().Net.Processor
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", net.Host, net.Port), thread.mux)
		if err != nil {
			thread.Interrupt <- threads.Panic
		}
	}(thread)
}

func (thread *Thread) Teardown() {

}
