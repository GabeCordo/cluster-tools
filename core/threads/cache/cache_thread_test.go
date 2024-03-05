package cache

import (
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"testing"
)

func TestThread_IncomingSaveRequest(t *testing.T) {

	in := make(chan common.CacheRequest, 1)
	out := make(chan common.CacheResponse, 1)
	thread := GenerateTestCacheThread(in, out)
	thread.Setup()
	go thread.Start()

	request := common.CacheRequest{
		Action:     common.CacheSaveIn,
		Identifier: "test",
		Data:       "blob",
		Nonce:      1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
	}
}

func TestThread_IncomingLoadRequest(t *testing.T) {

	// preliminary requirement to pull saved data
	in := make(chan common.CacheRequest, 1)
	out := make(chan common.CacheResponse, 1)
	thread := GenerateTestCacheThread(in, out)
	thread.Setup()
	go thread.Start()

	value := "blob"

	request := common.CacheRequest{
		Action: common.CacheSaveIn,
		Data:   value,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
		return
	}

	// checking the saved data
	request2 := common.CacheRequest{
		Action:     common.CacheLoadFrom,
		Identifier: response.Identifier,
		Nonce:      1,
	}

	in <- request2

	response2 := <-out

	if !response2.Success {
		t.Errorf("no value based on key %s found\n", response.Identifier)
		return
	}

	if response2.Data != value {
		t.Errorf("expected cache to return %s but got %s\n", value, response.Data)
	}
}

func TestThread_IncomingSaveSwapRequest(t *testing.T) {

	// preliminary requirement to pull saved data
	in := make(chan common.CacheRequest, 1)
	out := make(chan common.CacheResponse, 1)
	thread := GenerateTestCacheThread(in, out)
	thread.Setup()
	go thread.Start()

	value := "blob"

	request := common.CacheRequest{
		Action: common.CacheSaveIn,
		Data:   value,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
		return
	}

	// checking the saved data
	request2 := common.CacheRequest{
		Action:     common.CacheLoadFrom,
		Identifier: response.Identifier,
		Nonce:      2,
	}
	in <- request2

	response2 := <-out

	if !response2.Success {
		t.Errorf("no value based on key %s found\n", response.Identifier)
		return
	}

	if response2.Data != value {
		t.Errorf("expected cache to return %s but got %s\n", value, response.Data)
		return
	}

	value2 := "boop"

	// swap the value of the data
	request3 := common.CacheRequest{
		Action:     common.CacheSaveIn,
		Identifier: response.Identifier,
		Data:       value2,
		Nonce:      3,
	}
	in <- request3

	response3 := <-out

	if !response3.Success {
		t.Errorf("could not swap value at identifier %s\n", response.Identifier)
		return
	}

	if response.Identifier != response3.Identifier {
		t.Error("swapping a value should not change the identifier")
		return
	}

	request4 := common.CacheRequest{
		Action:     common.CacheLoadFrom,
		Identifier: response.Identifier,
		Nonce:      4,
	}
	in <- request4

	response4 := <-out

	if !response4.Success {
		t.Error("could not load swapped value for verification")
	}

	if response4.Data != value2 {
		t.Errorf("expected swapped value to be %s but was %s\n", value2, response4.Data)
	}
}
