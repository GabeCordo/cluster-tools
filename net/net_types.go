package net

import (
	"ETLFramework/logger"
	"crypto/ecdsa"
	"net/http"
	"sync"
)

type Router func(http.ResponseWriter, *http.Request)

type Request struct {
	lambda string
	param  []string
	auth   struct {
		signature []byte
		nonce     int
	}
}

type Node struct {
	name   string
	port   string
	debug  bool
	status NodeStatus

	auth   *NodeAuth
	logger *logger.Logger

	mutex sync.Mutex
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
	logger  *logger.Logger
	mutex   sync.Mutex
}
