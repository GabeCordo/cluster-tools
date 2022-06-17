package net

import (
	"ETLFramework/logger"
)

func NewAuth(logger *logger.Logger) NodeAuth {
	trustedMap := make(map[string]*NodeEndpoint)
	// mutex is initialized implicitly by the struct
	return NodeAuth{Trusted: trustedMap, Logger: logger}
}

func (na NodeAuth) AddTrusted(ip string, ne *NodeEndpoint) bool {
	if ne == nil {
		return false
	}
	na.Mutex.Lock()
	defer na.Mutex.Unlock()

	if _, ok := na.Trusted[ip]; !ok {
		na.Trusted[ip] = ne
		return true
	} else {
		return false
	}
}

func (na NodeAuth) RemoveTrusted(ip string) error {
	na.Mutex.Lock()
	delete(na.Trusted, ip)
	na.Mutex.Unlock()
	return nil
}

func (na NodeAuth) ValidateSource(ip string, hash, sig []byte) bool {
	validFlag := false // by default, we will assume that the ip doesn't exist in the hash map
	if endpoint, ok := na.Trusted[ip]; ok {
		validFlag = endpoint.ValidateSource(hash, sig)
	}
	return validFlag
}