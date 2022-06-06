package main

import (
	"ETLFramework/core"
	"fmt"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	core.FormatResponse(w, http.StatusOK, "{\"value\":\"good\"}")
}

func hi(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi")
}

func main() {
	n := core.NewNode("main", 10000, true)
	n.Route("/", index, []string{"GET"}, false)
	n.Route("/hi", hi, []string{"GET"}, true)
	n.Start()
}
