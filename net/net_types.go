package net

import (
	"crypto/ecdsa"
	"net/http"
)

type Router func(http.ResponseWriter, *http.Request)

type RequestAuth struct {
	signature []byte
	nonce     int
}

type Request struct {
	lambda string
	param  []string
	auth   *RequestAuth
}

type Node struct {
	name   string
	port   string
	debug  bool
	status NodeStatus
	auth   *NodeAuth
	logger *NodeLogger
}

type ILogger interface {
	Log(template string, params ...interface{})
	Alert(template string, params ...interface{})
	Warning(template string, params ...interface{})
}

type NodeLogger struct {
	folder string
	style  LoggerOutput
	node   *Node
}

type Permission struct {
	get    bool
	post   bool
	pull   bool
	delete bool
}

type NodeEndpoint struct {
	name              string
	publicKey         *ecdsa.PublicKey
	globalPermissions *Permission
	localPermissions  map[string]*Permission
}

type NodeAuth struct {
	trusted map[string]*NodeEndpoint
}
