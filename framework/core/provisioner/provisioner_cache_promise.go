package provisioner

import (
	"github.com/GabeCordo/etl-light/core/threads"
	"time"
)

const (
	NumOfAllowedCacheMisses            = 1000
	DefaultCacheWaitTimeInMilliseconds = 100
)

type CacheResponsePromise struct {
	nonce   uint32
	channel chan threads.CacheResponse
}

func NewCacheResponsePromise(nonce uint32, channel chan threads.CacheResponse) *CacheResponsePromise {
	promise := new(CacheResponsePromise)
	promise.nonce = nonce
	promise.channel = channel
	return promise
}

func (promise CacheResponsePromise) Wait() (threads.CacheResponse, bool) {

	// in the background, check to see if the provisioner thread has received and saved the response from the cache
	// in a local map that relates the Nonce sent to the CacheRequest channel with the Nonce of the recieved CacheResponse
	var response threads.CacheResponse

	go func() {
		time.Sleep(DefaultCacheWaitTimeInMilliseconds * time.Millisecond)
		GetProvisionerMemoryInstance().SendCacheResponseEvent(promise.nonce, threads.CacheResponse{Success: false})
	}()

	response = <-promise.channel

	// (CacheResponse, DidTimeoutOrFailureOccur)
	return response, !response.Success
}
