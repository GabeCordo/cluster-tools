package net

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"
)

func NewEndpoint(name string, publicKey *ecdsa.PublicKey) *Endpoint {
	endpoint := new(Endpoint)

	endpoint.Name = name
	endpoint.PublicKey = publicKey
	endpoint.LastNonce = MissingNonceValue
	endpoint.LocalPermissions = make(map[string]Permission)

	return endpoint
}

func (ne *Endpoint) AddGlobalPermission(permission Permission) {
	ne.GlobalPermissions = permission
}

func (ne *Endpoint) AddLocalPermission(route string, permission Permission) bool {
	if _, found := ne.LocalPermissions[route]; !found {
		ne.LocalPermissions[route] = permission
		return true
	}
	return false
}

func (ne *Endpoint) GetPublicKey() (*ecdsa.PublicKey, bool) {
	if (len(ne.X509) == 0) && (ne.PublicKey == nil) {
		return nil, false
	}

	if ne.PublicKey != nil {
		return ne.PublicKey, true
	}

	publicKeyByteArray, ok := StringToByte(ne.X509)
	if !ok {
		return nil, false
	}
	ne.GeneratePublicKey(publicKeyByteArray)

	return ne.PublicKey, true
}

func (ne *Endpoint) GeneratePublicKey(data []byte) bool {
	ne.X509 = ByteToString(data)

	// any ECDSA key stored in a byte format should be encoded using the x509 scheme
	// rather than the default ecdsa.Marshal encoding scheme
	publicKey, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return false
	}

	ne.PublicKey = publicKey.(*ecdsa.PublicKey)

	return true
}

func (ne *Endpoint) PublicKeyToBytes() []byte {
	if ne.PublicKey == nil {
		return []byte{}
	}

	b := elliptic.Marshal(ne.PublicKey.Curve, ne.PublicKey.X, ne.PublicKey.Y)
	return b
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
	} else {
		return ne.GlobalPermissions.Check(method)
	}
}

func (ne Endpoint) String() string {
	j, _ := json.Marshal(ne)
	return string(j)
}
