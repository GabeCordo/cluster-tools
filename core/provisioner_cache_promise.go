package core

import (
	"time"
)

const (
	NumOfAllowedCacheMisses            = 1000
	DefaultCacheWaitTimeInMilliseconds = 100
)

type CacheResponsePromise struct {
	nonce   uint32
	channel chan CacheResponse
}

func NewCacheResponsePromise(nonce uint32, channel chan CacheResponse) *CacheResponsePromise {
	promise := new(CacheResponsePromise)
	promise.nonce = nonce
	promise.channel = channel
	return promise
}

func (promise CacheResponsePromise) Wait() (CacheResponse, bool) {

	// in the background, check to see if the provisioner thread has received and saved the response from the cache
	// in a local map that relates the Nonce sent to the CacheRequest channel with the Nonce of the recieved CacheResponse
	var response CacheResponse

	go func() {
		time.Sleep(DefaultCacheWaitTimeInMilliseconds * time.Millisecond)
		GetProvisionerMemoryInstance().LinkCacheResponse(promise.nonce, CacheResponse{Success: false})
	}()

	response = <-promise.channel

	// (CacheResponse, DidTimeoutOrFailureOccur)
	return response, !response.Success
}
