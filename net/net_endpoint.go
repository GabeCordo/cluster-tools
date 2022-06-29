package net

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
)

func NewEndpoint(name string, publicKey *ecdsa.PublicKey) *Endpoint {
	endpoint := new(Endpoint)

	endpoint.Name = name
	endpoint.PublicKey = publicKey
	endpoint.LastNonce = MissingNonceValue
	endpoint.GlobalPermissions = nil
	endpoint.LocalPermissions = make(map[string]*Permission)

	return endpoint
}

func (ne *Endpoint) AddGlobalPermission(permission *Permission) {
	if permission != nil {
		ne.GlobalPermissions = permission
	}
}

func (ne *Endpoint) AddLocalPermission(route string, permission *Permission) {
	if permission != nil {
		ne.LocalPermissions[route] = permission
	}
}

func (ne *Endpoint) GeneratePublicKey(data []byte) {
	key := new(ecdsa.PublicKey)

	x, y := elliptic.Unmarshal(key.Curve, data)
	key.X = x
	key.Y = y

	ne.PublicKey = key
}

func (ne *Endpoint) PublicKeyToBytes() []byte {
	if ne.PublicKey == nil {
		return []byte{}
	}
	return elliptic.Marshal(ne.PublicKey.Curve, ne.PublicKey.X, ne.PublicKey.Y)
}

func (ne *Endpoint) ValidateSource(request *Request) bool {
	// if we do not have a public key we cannot verify the ECDSA signature
	if ne.PublicKey == nil {
		return false
	}
	// we cannot accept the last received or previous nonce, or we risk a threat actor
	// resending an intercepted nonce/signature to forge credentials
	if request.Auth.Nonce <= ne.LastNonce {
		return false
	}
	hash := request.Hash()
	return ecdsa.VerifyASN1(ne.PublicKey, hash[:], request.Auth.Signature)
}

func (ne Endpoint) HasPermissionToUseMethod(route, method string) bool {
	if localPermission, ok := ne.LocalPermissions[route]; ok {
		return localPermission.Check(method)
	} else if ne.GlobalPermissions != nil {
		return ne.GlobalPermissions.Check(method)
	} else {
		return false
	}
}

func (ne Endpoint) String() string {
	j, _ := json.Marshal(ne)
	return string(j)
}
