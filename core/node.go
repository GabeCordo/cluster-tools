package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// ~

type NodeStatus int

const (
	Startup NodeStatus = iota
	Running
	Frozen
	Killed
)

// ~

type NodeIllegalActionError struct{}

func (e *NodeIllegalActionError) Error() string {
	return "Illegal Action Given the Node's Current State"
}

// ~

type Router func(http.ResponseWriter, *http.Request)

// ~

type Request struct {
	nonce     int
	signature []byte
	data      map[string]string
}

// ~

type INode interface{}

type Node struct {
	name   string
	port   string
	debug  bool
	status NodeStatus
	auth   *NodeAuth
	logger *NodeLogger
}

func NewNode(name string, port int, debug bool, logger *NodeLogger) Node {
	portString := fmt.Sprintf(":%d", port)
	authInstance := NewAuth()
	if logger == nil {
		return Node{name, portString, debug, Startup, &authInstance, logger}
	} else {
		return Node{name, portString, debug, Startup, &authInstance, nil}
	}
}

func (n Node) Route(path string, handler Router, methods []string, auth bool) error {
	if n.status == Startup {
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if !IsUsingJSONContent(r) {
				http.Error(w, "Only JSON Content permitted", http.StatusBadRequest)
				return
			}
			flag := false
			for _, m := range methods {
				if m == r.Method {
					flag = true
				}
			}
			if flag {
				fmt.Println("Request Good")
				var body Request
				err := json.NewDecoder(r.Body).Decode(&body)
				if err != nil {
					fmt.Println("Decoder Error")
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				if auth {
					ip := GetInternetProtocol(r)
					hash := sha256.Sum256([]byte(path + strconv.Itoa(body.nonce)))
					if n.auth.ValidateSource(ip, hash[:], body.signature) {
						handler(w, r)
					} else {
						http.Error(w, "blah", http.StatusForbidden)
					}
				} else {
					handler(w, r)
				}
			} else {
				strError := fmt.Sprintf("%s not allowed for /%s", r.Method, path)
				n.logger.Log(strError)
			}
		})
		return nil
	}
	return &NodeIllegalActionError{}
}

func (n Node) Start() {
	n.status = Running
	n.logger.Log("Starting up ETLNode Server over port ")
	log.Fatal(http.ListenAndServe(n.port, nil))
}

func GetInternetProtocol(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func FormatResponse(w http.ResponseWriter, httpResponseCode int, data string) string {
	response := fmt.Sprintf("{\"status\":%d, \"data\": %s }", httpResponseCode, data)
	fmt.Println(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpResponseCode)
	json.NewEncoder(w).Encode(response)
	return response
}

func IsUsingJSONContent(r *http.Request) bool {
	content := r.Header.Get("Content-Type")
	return content == "application/json"
}
