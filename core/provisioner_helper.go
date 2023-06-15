package core

import (
	"errors"
	"math/rand"
)

type Helper struct {
	c9  chan CacheRequest
	c11 chan MessengerRequest
}

func NewHelper(channels ...any) (*Helper, error) {
	helper := new(Helper)

	if len(channels) < 2 {
		return nil, errors.New("helper requires two channels")
	}

	var ok bool // tracks the state of the channels provided

	helper.c9, ok = (channels[0]).(chan CacheRequest)
	if !ok {
		return nil, errors.New("first parameter must be channel C9")
	}

	helper.c11, ok = (channels[1]).(chan MessengerRequest)
	if !ok {
		return nil, errors.New("second parameter must be channel C11")
	}

	return helper, nil
}

func (helper Helper) IsDebugEnabled() bool {
	return GetConfigInstance().Debug
}

func (helper Helper) SaveToCache(data any) *CacheResponsePromise {

	var expiry float64
	if GetConfigInstance().Cache.Expiry != 0.0 {
		expiry = GetConfigInstance().Cache.Expiry
	} else {
		expiry = DefaultTimeout
	}

	requestNonce := rand.Uint32()
	helper.c9 <- CacheRequest{Action: CacheSaveIn, Data: data, Nonce: requestNonce, ExpiresIn: expiry}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) LoadFromCache(identifier string) *CacheResponsePromise {

	requestNonce := rand.Uint32()
	helper.c9 <- CacheRequest{Action: CacheLoadFrom, Identifier: identifier, Nonce: requestNonce}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) Log(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- MessengerRequest{Action: MessengerLog, Cluster: cluster, Message: message, Nonce: requestNonce}
}

func (helper Helper) Warning(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- MessengerRequest{Action: MessengerWarning, Cluster: cluster, Message: message, Nonce: requestNonce}
}

func (helper Helper) Fatal(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- MessengerRequest{Action: MessengerFatal, Cluster: cluster, Message: message, Nonce: requestNonce}
}
