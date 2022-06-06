package main

import (
	"ETLFramework/core"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

/*? Routing Functions */

func index(w http.ResponseWriter, r *http.Request) {
	core.FormatResponse(w, http.StatusOK, "{\"value\":\"good\"}")
}

func onlyPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi")
}

/*? Source Code to Test */

func StartupHTTPNode() {
	n := core.NewNode("main", 10000, false, nil) // pass a nil to logger pointer ~ no logging
	n.Route("/", index, []string{"GET"}, false)
	n.Route("/onlyPost", onlyPost, []string{"POST"}, true)
	go n.Start()
}

/*? Test Helper Code */

func GetHTTPResponseJSON(resp *http.Response) map[string]string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	bodyJson := map[string]string{}
	json.Unmarshal([]byte(body), &bodyJson)

	return bodyJson
}

/*? Test Function */

func TestNodeHTTP(t *testing.T) {
	StartupHTTPNode()
	resp, err := http.Get("localhost:10000/")
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}
	respJson := GetHTTPResponseJSON(resp)
	if respJson["value"] != "good" {
		t.Error("HTTP Get route returned the wrong data")
	}
}
