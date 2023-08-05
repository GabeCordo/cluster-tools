package provisioner

import (
	"errors"
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl-light/utils"
	"github.com/GabeCordo/etl/framework/core/common"
	"math/rand"
)

type Helper struct {
	c9  chan threads.CacheRequest
	c11 chan threads.MessengerRequest
}

func NewHelper(channels ...any) (utils.Helper, error) {
	helper := new(Helper)

	if len(channels) < 2 {
		return nil, errors.New("helper requires two channels")
	}

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
	helper.c9 <- threads.CacheRequest{Action: threads.CacheSaveIn, Data: data, Nonce: requestNonce, ExpiresIn: expiry}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) LoadFromCache(identifier string) utils.Promise {

	requestNonce := rand.Uint32()
	helper.c9 <- threads.CacheRequest{Action: threads.CacheLoadFrom, Identifier: identifier, Nonce: requestNonce}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) Log(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- threads.MessengerRequest{Action: threads.MessengerLog, Cluster: cluster, Message: message, Nonce: requestNonce}
}

func (helper Helper) Warning(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- threads.MessengerRequest{Action: threads.MessengerWarning, Cluster: cluster, Message: message, Nonce: requestNonce}
}

func (helper Helper) Fatal(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.c11 <- threads.MessengerRequest{Action: threads.MessengerFatal, Cluster: cluster, Message: message, Nonce: requestNonce}
}
