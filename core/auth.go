package core

import (
	"crypto/ecdsa"
)

// ~

type bitmap = [4]bool

// ~

type NodeEndpoint struct {
	name             string
	permissions      bitmap
	publicKey        ecdsa.PublicKey
	localPermissions map[string]bitmap
}

func NewEndpoint(name string, permissions bitmap, publicKey ecdsa.PublicKey) NodeEndpoint {
	tm := make(map[string]bitmap)
	return NodeEndpoint{name, permissions, publicKey, tm}
}

// ~

type NodeAuth struct {
	trusted map[string]*NodeEndpoint
}

func NewAuth() NodeAuth {
	m := make(map[string]*NodeEndpoint)
	return NodeAuth{m}
}

func (na NodeAuth) AddTrusted(ip string, ne *NodeEndpoint) error {
	if _, ok := na.trusted[ip]; !ok {
		na.trusted[ip] = ne
	}
	return nil
}

func (na NodeAuth) RemoveTrusted(ip string) error {
	delete(na.trusted, ip)
	return nil
}

func (na NodeAuth) ValidateSource(ip string, hash, sig []byte) bool {
	if endpoint, ok := na.trusted[ip]; ok {
		return ecdsa.VerifyASN1(&endpoint.publicKey, hash, sig)
	}
	return false
}
