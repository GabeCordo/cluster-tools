package net

import (
	"ETLFramework/logger"
	"crypto/ecdsa"
	"net/http"
	"sync"
)

const (
	localhost = "127.0.0.1"
)

type Address struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Router func(request *Request, response *Response)

type Request struct {
	Function string   `json:"function"`
	Param    []string `json:"param"`
	Auth     struct {
		Signature []byte `json:"signature"`
		Nonce     int64  `json:"nonce"`
	} `json:"auth"`
}

type ResponseData map[string]interface{}

type Response struct {
	Status      int          `json:"status"`
	Description string       `json:"description"`
	Data        ResponseData `json:"data"`
}

type Node struct {
	Name    string     `json:"name"`
	Address Address    `json:"address"`
	Debug   bool       `json:"debug"`
	Status  NodeStatus `json:"status"`

	Auth   *Auth
	Logger *logger.Logger

	mux    *http.ServeMux
	server *http.Server
	mutex  sync.Mutex
}

type Permission struct {
	Get    bool `json:"get"`
	Post   bool `json:"post"`
	Pull   bool `json:"pull"`
	Delete bool `json:"delete"`
}

type Endpoint struct {
	Name              string `json:"name"`
	X509              string `json:"publicKey"`
	PublicKey         *ecdsa.PublicKey
	LastNonce         int64
	GlobalPermissions Permission            `json:"globalPermissions"`
	LocalPermissions  map[string]Permission `json:"localPermissions"`
}

type Auth struct {
	Trusted map[string]*Endpoint `json:"trusted"`
	Mutex   sync.Mutex           `json:"mutex"`
}
