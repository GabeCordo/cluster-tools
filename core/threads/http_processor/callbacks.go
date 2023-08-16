package http_processor

import (
	"net/http"
)

func (thread *Thread) processorCallback(w http.ResponseWriter, r *http.Request) {

}

func (thread *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

}

func (thread *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		// treat this as a probe to the server
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
