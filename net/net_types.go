package net

import (
	"ETLFramework/logger"
	"crypto/ecdsa"
	"net/http"
	"sync"
)

type Address struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Router func(http.ResponseWriter, *http.Request)

type Request struct {
	Lambda string   `json:"lambda"`
	Param  []string `json:"param"`
	Auth   struct {
		Signature []byte `json:"signature"`
		Nonce     int    `json:"nonce"`
	} `json:"auth"`
}

type Node struct {
	Name    string     `json:"name"`
	Address Address    `json:"address"`
	Debug   bool       `json:"debug"`
	Status  NodeStatus `json:"status"`

	Auth   *NodeAuth
	Logger *logger.Logger

	Mutex sync.Mutex
}

type Permission struct {
	Get    bool `json:"get"`
	Post   bool `json:"post"`
	Pull   bool `json:"pull"`
	Delete bool `json:"delete"`
}

type NodeEndpoint struct {
	Name              string                 `json:"name"`
	PublicKey         *ecdsa.PublicKey       `json:"publicKey"`
	GlobalPermissions *Permission            `json:"globalPermissions"`
	LocalPermissions  map[string]*Permission `json:"localPermissions"`
}

type NodeAuth struct {
	Trusted map[string]*NodeEndpoint `json:"trusted"`
	Logger  *logger.Logger           `json:"logger"`
	Mutex   sync.Mutex               `json:"mutex"`
}
