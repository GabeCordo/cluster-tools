package core

import (
	"sync"
	"time"
)

type CacheResponsePromise struct {
	nonce  uint32
	record CacheResponse
	wg     sync.WaitGroup
}

func NewCacheResponsePromise() *CacheResponsePromise {
	promise := new(CacheResponsePromise)
	return promise
}

func (cacheResponsePromise CacheResponsePromise) Wait() CacheResponse {

	// in the background, check to see if the provisioner thread has received and saved the response from the cache
	// in a local map that relates the Nonce sent to the CacheRequest channel with the Nonce of the recieved CacheResponse
	go func() {
		for {
			if data, found := GetProvisionerMemoryInstance().PopCacheResponse(cacheResponsePromise.nonce); found {
				// the record is initially set to nil to signify that the data hasn't been received yet
				// so now that we have grabbed it from the provisioner response cache, we can copy it to the record
				// and delete it from the provisioner cache
				cacheResponsePromise.record = data
				GetProvisionerMemoryInstance().cache.Delete(cacheResponsePromise.nonce)

				// when we create a request to search for the data, we increment the work group by one, we can
				// mark it as done to allow the Wait function to continue and return to the calling code block
				cacheResponsePromise.wg.Done()
				break
			}
			time.Sleep(1 * time.Microsecond)
		}
	}()

	// Block until a CacheRequest with the same Nonce as the CacheResponse is received
	cacheResponsePromise.wg.Wait()

	// Let some work be performed with the record
	return cacheResponsePromise.record
}
