package api

import "net/http"

func IsDebugEnabled(host string) bool {

	rsp, err := http.Get(host + "/debug")
	if err != nil {
		return false
	}
	defer rsp.Body.Close()

	return rsp.Status == "200 OK"
}
