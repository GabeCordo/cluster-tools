package provisioner

import (
	"errors"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl-light/utils"
	"github.com/GabeCordo/etl/core/threads/common"
	"math/rand"
)

type Helper struct {
	module  string
	cluster string
	c9      chan threads.CacheRequest
	c11     chan threads.MessengerRequest
}

func NewHelper(module, cluster string, channels ...any) (*Helper, error) {
	helper := new(Helper)

	if len(channels) < 2 {
		return nil, errors.New("helper requires two channels")
	}

	helper.module = module
	helper.cluster = cluster

	var ok bool // tracks the state of the channels provided

	helper.c9, ok = (channels[0]).(chan threads.CacheRequest)
	if !ok {
		return nil, errors.New("first parameter must be channel C9")
	}

	helper.c11, ok = (channels[1]).(chan threads.MessengerRequest)
	if !ok {
		return nil, errors.New("second parameter must be channel C11")
	}

	return helper, nil
}

func (helper Helper) IsDebugEnabled() bool {
	return common.GetConfigInstance().Debug
}

func (helper Helper) SaveToCache(data any) utils.Promise {

	var expiry float64
	if common.GetConfigInstance().Cache.Expiry != 0.0 {
		expiry = common.GetConfigInstance().Cache.Expiry
	} else {
		expiry = threads.DefaultTimeout
	}

	requestNonce := rand.Uint32()
	helper.c9 <- threads.CacheRequest{
		Action:    threads.CacheSaveIn,
		Data:      data,
		Nonce:     requestNonce,
		ExpiresIn: expiry,
	}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) LoadFromCache(identifier string) utils.Promise {

	requestNonce := rand.Uint32()
	helper.c9 <- threads.CacheRequest{
		Action:     threads.CacheLoadFrom,
		Identifier: identifier,
		Nonce:      requestNonce,
	}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) Log(message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- threads.MessengerRequest{
		Action:  threads.MessengerLog,
		Module:  helper.module,
		Cluster: helper.cluster,
		Message: message,
		Nonce:   requestNonce,
	}
}

func (helper Helper) Warning(message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- threads.MessengerRequest{
		Action:  threads.MessengerWarning,
		Module:  helper.module,
		Cluster: helper.cluster,
		Message: message,
		Nonce:   requestNonce,
	}
}

func (helper Helper) Fatal(message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- threads.MessengerRequest{
		Action:  threads.MessengerFatal,
		Module:  helper.module,
		Cluster: helper.cluster,
		Message: message,
		Nonce:   requestNonce,
	}
}
