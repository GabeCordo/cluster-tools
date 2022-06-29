package net

import (
	"ETLFramework/logger"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

// NewNode
// address : Address -> defines the listening host and port
// auth : *Auth ->
//
// LEGACY: func NewNode(address Address, debug bool, auth *Auth, logger *logger.Logger) *Node {
func NewNode(address Address, optional ...interface{}) *Node {
	node := new(Node)

	for _, o := range optional {
		switch val := o.(type) {
		case *Auth:
			node.Auth = val // default: nil
		case *logger.Logger:
			node.Logger = val // default: nil
		case bool:
			node.Debug = val // default: false
		}
	}

	node.Name = GenerateRandomString(int(GenerateNonce()))
	node.Mux = http.NewServeMux()
	node.Address = address
	node.Status = Startup

	return node
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

func (n *Node) SetStatus(status NodeStatus) {
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	n.Status = status // if we don't lock this, two threads attempting to change the status can cause a race condition
}

func (n *Node) SetName(name string) {
	// a node name should be static during runtime given that during a time interval from (0 to inf)
	// if N logs are stored, using the node name as an id, then a dynamic name change at time t
	// would render new logs created from (t to inf) detached from logs created from (0 to t)
	if n.Status == Startup {
		n.Name = name
	}
}

func (n *Node) Function(path string, handler Router, methods []string, auth bool) error {

	// the user should not be able to create a route that requires ECDSA or permission bitmap
	// authentication if they have not registered an auth structure
	if (n.Auth == nil) && auth {
		return errors.New("cannot create a route that requires authentication with a nil auth struct")
	}

	// the design pattern calls for routes to be appended before the HTTP server is started up
	// so ensure that it cannot be called during runtime
	if n.Status == Startup {

		// by passing a hash-map to the handler function instead of a list of methods, we can perform a method
		// look-up in O(1) time versus O(n) needed to iterate over a list of methods
		methodsHashTable := ArrayToLookupHashTable(methods)

		n.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()

			response := NewResponse()
			defer response.Send(w)

			HTTPRequestAllowedOnHandler := false
			if _, ok := methodsHashTable[r.Method]; ok {
				HTTPRequestAllowedOnHandler = true
			}

			if HTTPRequestAllowedOnHandler {
				if n.Logger != nil {
					n.Logger.Log(n.Name, "Request %s used an approved %s method.", r.Host, r.Method)
				}

				if !IsUsingJSONContent(r) {
					response.AddStatus(http.StatusBadRequest, "Only JSON Content permitted")
					return
				}

				// we will see if the IP address has a mapped local or global permission to the endpoint
				sender, error := GetInternetProtocol(r)
				if error != nil {
					response.AddStatus(http.StatusInternalServerError, "Internet Protocol Parser Failed")
					return
				}

				/** Unmarshal the JSON body to Request Struct */
				body := new(Request)

				httpBodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					response.AddStatus(http.StatusInternalServerError, err.Error())
					return
				}
				s := string(httpBodyBytes)
				err = json.Unmarshal([]byte(s), body)
				if err != nil {
					if n.Logger != nil {
						n.Logger.Log(n.Name, "Request %s contained a malformed HTTP body", r.Host)
					}
					response.AddStatus(http.StatusBadRequest, err.Error())
					return
				}

				if auth {
					// Why not pass the lambda provided by the request to IsEndpointAuthorized?
					//		-> the user is not forced to use the request.Send() method and can
					//		   direct the request to an url they do not have permission for while
					//		   inserting an url path as the lambda for a route they do have permission
					//		   for
					// Why not place method into request type as well?
					//		-> a lambda can support > 1 HTTP method
					//		-> it is safer to use a server-defined method that the node has control over
					if n.Auth.IsEndpointAuthorized(sender, body, path, r.Method) {
						// the request IP destination either had local or global permission
						handler(response)
					} else {
						// the request IP destination does not have local or global permission
						if n.Logger != nil {
							n.Logger.Warning("Request %s attempted to submit a request to %s (%s); did not have permission", path, r.Method)
						}
						response.AddStatus(http.StatusUnauthorized, "Bye Bye.")
					}
				} else {
					// the endpoint does not require the destination ip of the request to have local or global
					// permission to send messages to the Node
					handler(response)
				}
			} else {
				if n.Logger != nil {
					n.Logger.Log(n.Name, "Request to %s failed; Path does not support %s", path, r.Method)
				}
				response.AddStatus(http.StatusForbidden, "HTTP Method Not Allowed")
			}

			// this exists in the event that an unintended error or unforeseen error has been improperly handled
			// by a user-defined handler function or a packet has corrupted IP / body data
			defer func() {
				if err := recover(); err != nil {
					response.AddStatus(http.StatusInternalServerError, "Node panic")
				}
			}()
		})
		return nil
	}
	// the developer should know that they're breaking the pattern by calling this during runtime
	return &NodeIllegalActionError{}
}

func (n *Node) Start() {
	n.SetStatus(Running) // thread safe
	if n.Logger != nil {
		n.Logger.Log(n.Name, "starting up Node(%s) on %s:%d", n.Name, n.Address.Host, n.Address.Port)
	}
	http.ListenAndServe(n.Address.String(), n.Mux)
}

func (n Node) String() string {
	j, _ := json.Marshal(n)
	return string(j)
}
