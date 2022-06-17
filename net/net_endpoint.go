package net

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
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

	ne.PublicKey = key
}

func (ne NodeEndpoint) PublicKeyToBytes() []byte {
	if ne.PublicKey == nil {
		return []byte{}
	}
	return elliptic.Marshal(ne.PublicKey.Curve, ne.PublicKey.X, ne.PublicKey.Y)
}

func (ne NodeEndpoint) ValidateSource(hash, sig []byte) bool {
	if ne.PublicKey == nil {
		return false
	}
	return ecdsa.VerifyASN1(ne.PublicKey, hash, sig)
}

func (ne NodeEndpoint) HasPermission(route string, method string) bool {
	if localPermission, ok := ne.LocalPermissions[route]; ok {
		return localPermission.Check(method)
	} else {
		return ne.GlobalPermissions.Check(method)
	}
}

func (ne NodeEndpoint) String() string {
	j, _ := json.Marshal(ne)
	return string(j)
}
