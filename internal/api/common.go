package api

import "net/http"

var client = http.Client{}

type Response struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}
