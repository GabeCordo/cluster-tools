package net

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
		return endpoint.ValidateSource(hash, sig)
	}
	return false
}
