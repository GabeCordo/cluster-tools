package net

import (
	"crypto/ecdsa"
	"crypto/elliptic"
)

func NewEndpoint(name string, permissions *Permission, publicKey *ecdsa.PublicKey) NodeEndpoint {
	localPermissions := make(map[string]*Permission)
	return NodeEndpoint{name, publicKey, permissions, localPermissions}
}

func (ne NodeEndpoint) GeneratePublicKey(data []byte) {
	key := new(ecdsa.PublicKey)

	x, y := elliptic.Unmarshal(key.Curve, data)
	key.X = x
	key.Y = y

	ne.publicKey = key
}

func (ne NodeEndpoint) PublicKeyToBytes() []byte {
	if ne.publicKey == nil {
		return []byte{}
	}
	return elliptic.Marshal(ne.publicKey.Curve, ne.publicKey.X, ne.publicKey.Y)
}

func (ne NodeEndpoint) ValidateSource(hash, sig []byte) bool {
	if ne.publicKey == nil {
		return false
	}
	return ecdsa.VerifyASN1(ne.publicKey, hash, sig)
}

func (ne NodeEndpoint) HasPermission(route string, method string) bool {
	if localPermission, ok := ne.localPermissions[route]; ok {
		return localPermission.Check(method)
	} else {
		return ne.globalPermissions.Check(method)
	}
}
