package net

import (
	"encoding/json"
	"net/http"
)

func NewResponse() *Response {
	response := new(Response)
	response.Status = http.StatusNoContent // if the status is never populated, this will be returned by the node
	response.Data = make(ResponseData)
	return response
}

func (r *Response) SetStatus(httpResponseCode int) {
	r.Status = httpResponseCode
}

func (r *Response) AddPair(key string, value interface{}) {
	r.Data[key] = value
}

func (r *Response) AddStatus(httpResponseCode int, message string) {
	r.Status = httpResponseCode
	r.Data["status"] = message
}

func (r *Response) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)
	json.NewEncoder(w).Encode(r)
}
