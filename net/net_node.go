package net

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type NodeStatus int

const (
	Startup NodeStatus = iota
	Running
	Frozen
	Killed
)

func (ns NodeStatus) String() string {
	switch ns {
	case Startup:
		return "Startup"
	case Running:
		return "Running"
	case Frozen:
		return "Frozen"
	default:
		return "Killed"
	}
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

func (n Node) IsAuthAttached() bool {
	return n.auth == nil
}

func (n Node) IsLoggerAttached() bool {
	return n.logger == nil
}

func (n Node) DebugMode() bool {
	return n.IsLoggerAttached() && n.debug
}

func (n Node) Route(path string, handler Router, methods []string, auth bool) error {
	if n.status == Startup {
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if !IsUsingJSONContent(r) {
				http.Error(w, "Only JSON Content permitted", http.StatusBadRequest)
				return
			}
			flagRequestToAllowedMethod := false
			for _, method := range methods {
				if method == r.Method {
					flagRequestToAllowedMethod = true
				}
			}
			if flagRequestToAllowedMethod {
				fmt.Println("Request Good")

				/** Unmarshal the JSON body to Request Struct */
				var body Request
				err := json.NewDecoder(r.Body).Decode(&body)
				if err != nil {
					fmt.Println("Decoder Error")
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				if auth {
					ip := GetInternetProtocol(r)
					hash := sha256.Sum256([]byte(path + strconv.Itoa(body.auth.nonce)))
					if n.auth.ValidateSource(ip, hash[:], body.auth.signature) {
						handler(w, r)
					} else {
						http.Error(w, "blah", http.StatusForbidden)
					}
				} else {
					handler(w, r)
				}
			} else {
				strError := fmt.Sprintf("%s not allowed for /%s", r.Method, path)
				if n.logger != nil {
					n.logger.Log(strError)
				}
			}
		})
		return nil
	}
	return &NodeIllegalActionError{}
}

func (n Node) Start() {
	n.status = Running
	if n.logger != nil {
		n.logger.Log("Starting up ETLNode Server over port ")
	}
	log.Fatal(http.ListenAndServe(n.port, nil))
}

func (n Node) String() string {
	j, _ := json.Marshal(n)
	return string(j)
}
