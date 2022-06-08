package net

import (
	"ETLFramework/logger"
)

func NewAuth(logger *logger.Logger) NodeAuth {
	trustedMap := make(map[string]*NodeEndpoint)
	// mutex is initialized implicitly by the struct
	return NodeAuth{trusted: trustedMap, logger: logger}
}

func (na NodeAuth) AddTrusted(ip string, ne *NodeEndpoint) bool {
	if ne == nil {
		return false
	}
	na.mutex.Lock()
	defer na.mutex.Unlock()

	if _, ok := na.trusted[ip]; !ok {
		na.trusted[ip] = ne
		return true
	} else {
		return false
	}
}

func (na NodeAuth) RemoveTrusted(ip string) error {
	na.mutex.Lock()
	delete(na.trusted, ip)
	na.mutex.Unlock()
	return nil
}

func (na NodeAuth) ValidateSource(ip string, hash, sig []byte) bool {
	validFlag := false // by default, we will assume that the ip doesn't exist in the hash map
	if endpoint, ok := na.trusted[ip]; ok {
		validFlag = endpoint.ValidateSource(hash, sig)
	}
	return validFlag
}
