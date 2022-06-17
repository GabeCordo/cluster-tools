package net

import (
	"ETLFramework/logger"
	"crypto/sha256"
	"encoding/json"
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

func NewNode(name string, address Address, debug bool, auth *NodeAuth, logger *logger.Logger) Node {
	return Node{Name: name, Address: address, Debug: debug, Status: Startup, Auth: auth, Logger: logger}
}

func (n Node) IsAuthAttached() bool {
	return n.Auth == nil
}

func (n Node) IsLoggerAttached() bool {
	return n.Logger == nil
}

func (n Node) MissingModules() bool {
	return (n.Auth == nil) || (n.Logger == nil)
}

func (n Node) SetStatus(status NodeStatus) {
	n.Mutex.Lock()
	n.Status = status // if we don't lock this, two threads attempting to change the status can cause a race condition
	n.Mutex.Unlock()
}

func (n Node) Route(path string, handler Router, methods []string, auth bool) error {
	// the design pattern calls for routes to be appended before the HTTP server is started up
	// so ensure that it cannot be called during runtime
	if n.Status == Startup {
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
				n.Logger.Log(n.Name, "Request %s used an approved %s method.", r.Host, r.Method)

				/** Unmarshal the JSON body to Request Struct */
				var body Request
				err := json.NewDecoder(r.Body).Decode(&body)
				if err != nil {
					n.Logger.Log(n.Name, "Request %s contained a malformed HTTP body", r.Host)
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				if auth {
					ip := GetInternetProtocol(r)
					hash := sha256.Sum256([]byte(path + strconv.Itoa(body.Auth.Nonce)))
					if n.Auth.ValidateSource(ip, hash[:], body.Auth.Signature) {
						handler(w, r)
					} else {
						n.Logger.Warning("Request %s attempted to submit a request to %s (%s); did not have permission", path, r.Method)
						http.Error(w, "Bye Bye.", http.StatusForbidden)
					}
				} else {
					// we do not know what handler may be passed to this function while registering a route,
					// run it as a go thread to avoid blocking the main execution thread
					go handler(w, r)
				}
			} else {
				if n.Logger != nil {
					n.Logger.Log(n.Name, "Request to %s failed; Path does not support %s", path, r.Method)
				}
			}
		})
		return nil
	}
	return &NodeIllegalActionError{} // the developer should know that they're breaking
	// the pattern by calling this during runtime
}

func (n Node) Start() {
	n.SetStatus(Running) // thread safe
	if n.Logger != nil {
		n.Logger.Log(n.Name, "Starting up ETLNode Server over port ")
	}
	log.Fatal(http.ListenAndServe(n.Address.String(), nil))
}

func (n Node) String() string {
	j, _ := json.Marshal(n)
	return string(j)
}
