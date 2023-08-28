package api

import "net/http"

func IsDebugEnabled(host string) bool {

	_, err := http.Get(host + "/debug")
	return err == nil
}
