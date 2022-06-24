package net

func NewAuth() *Auth {
	// mutex is initialized implicitly by the struct
	auth := new(Auth)
	auth.Trusted = make(map[string]*Endpoint)
	return auth
}

func (na *Auth) AddTrusted(ip string, ne *Endpoint) bool {
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

func (na *Auth) RemoveTrusted(ip string) error {
	na.Mutex.Lock()
	delete(na.Trusted, ip)
	na.Mutex.Unlock()
	return nil
}

func (na *Auth) IsEndpointAuthorized(sender *Address, request *Request, path, method string) bool {
	validFlag := false // by default, we will assume that the ip doesn't exist in the hash map
	if endpoint, ok := na.Trusted[sender.Host]; ok {
		// 1. does the user have permission to send an HTTP method request to the current path
		// 2. does the message come from a user with the same ECDSA key pair
		validFlag = endpoint.HasPermissionToUseMethod(path, method) && endpoint.ValidateSource(request)
	}
	return validFlag
}
