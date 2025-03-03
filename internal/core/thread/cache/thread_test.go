package cache

import (
	"github.com/Sentmint/pops/internal/core/thread"
	"testing"
)

func TestThread_IncomingSaveRequest(t *testing.T) {

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)
	th := GenerateTestCacheThread(in, out)
	th.Setup()
	go th.Start()

	request := thread.Request{
		Action: thread.CreateAction,
		Data:   thread.CacheRequestData{Identifier: "test", Data: "blob"},
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
	}
}

func TestThread_IncomingLoadRequest(t *testing.T) {

	// preliminary requirement to pull saved data
	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)
	th := GenerateTestCacheThread(in, out)
	th.Setup()
	go th.Start()

	identifier := "test"
	value := "blob"

	request := thread.Request{
		Action: thread.CreateAction,
		Data:   thread.CacheRequestData{Identifier: identifier, Data: value},
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
		return
	}

	cacheResponseData := (response.Data).(thread.CacheResponseData)

	// checking the saved data
	request2 := thread.Request{
		Action: thread.GetAction,
		Data:   thread.CacheRequestData{Identifier: cacheResponseData.Identifier},
		Nonce:  1,
	}

	in <- request2

	response2 := <-out

	cacheResponseData2 := (response2.Data).(thread.CacheResponseData)

	if !response2.Success {
		t.Errorf("no value based on key %s found\n", cacheResponseData2.Identifier)
		return
	}

	if cacheResponseData2.Data != value {
		t.Errorf("expected cache to return %s but got %s\n", value, response.Data)
	}
}

func TestThread_IncomingSaveSwapRequest(t *testing.T) {

	// preliminary requirement to pull saved data
	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)
	th := GenerateTestCacheThread(in, out)
	th.Setup()
	go th.Start()

	value := "blob"

	request := thread.Request{
		Action: thread.CreateAction,
		Data:   thread.CacheRequestData{Identifier: "test", Data: value},
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
		return
	}

	cacheResponseData := (response.Data).(thread.CacheResponseData)

	// checking the saved data
	request2 := thread.Request{
		Action: thread.GetAction,
		Data:   thread.CacheRequestData{Identifier: cacheResponseData.Identifier},
		Nonce:  2,
	}
	in <- request2

	response2 := <-out

	cacheResponseData2 := (response2.Data).(thread.CacheResponseData)

	if !response2.Success {
		t.Errorf("no value based on key %s found\n", cacheResponseData2.Identifier)
		return
	}

	if cacheResponseData2.Data != value {
		t.Errorf("expected cache to return %s but got %s\n", value, cacheResponseData2.Data)
		return
	}

	value2 := "boop"

	// swap the value of the data
	request3 := thread.Request{
		Action: thread.CreateAction,
		Data:   thread.CacheRequestData{Identifier: cacheResponseData.Identifier, Data: value2},
		Nonce:  3,
	}
	in <- request3

	response3 := <-out

	cacheResponseData3 := (response3.Data).(thread.CacheResponseData)

	if !response3.Success {
		t.Errorf("could not swap value at identifier %s\n", cacheResponseData.Identifier)
		return
	}

	if cacheResponseData.Identifier != cacheResponseData3.Identifier {
		t.Error("swapping a value should not change the identifier")
		return
	}

	request4 := thread.Request{
		Action: thread.GetAction,
		Data:   thread.CacheRequestData{Identifier: cacheResponseData.Identifier},
		Nonce:  4,
	}
	in <- request4

	response4 := <-out

	cacheResponseData4 := (response4.Data).(thread.CacheResponseData)

	if !response4.Success {
		t.Error("could not load swapped value for verification")
	}

	if cacheResponseData4.Data != value2 {
		t.Errorf("expected swapped value to be %s but was %s\n", value2, cacheResponseData4.Data)
	}
}
