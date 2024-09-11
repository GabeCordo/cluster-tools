package cache

import (
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"testing"
)

func TestThread_IncomingSaveRequest(t *testing.T) {

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)
	thread := GenerateTestCacheThread(in, out)
	thread.Setup()
	go thread.Start()

	request := common.ThreadRequest{
		Action: common.CreateAction,
		Data:   common.CacheRequestData{Identifier: "test", Data: "blob"},
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
	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)
	thread := GenerateTestCacheThread(in, out)
	thread.Setup()
	go thread.Start()

	identifier := "test"
	value := "blob"

	request := common.ThreadRequest{
		Action: common.CreateAction,
		Data:   common.CacheRequestData{Identifier: identifier, Data: value},
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
		return
	}

	cacheResponseData := (response.Data).(common.CacheResponseData)

	// checking the saved data
	request2 := common.ThreadRequest{
		Action: common.GetAction,
		Data:   common.CacheRequestData{Identifier: cacheResponseData.Identifier},
		Nonce:  1,
	}

	in <- request2

	response2 := <-out

	cacheResponseData2 := (response2.Data).(common.CacheResponseData)

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
	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)
	thread := GenerateTestCacheThread(in, out)
	thread.Setup()
	go thread.Start()

	value := "blob"

	request := common.ThreadRequest{
		Action: common.CreateAction,
		Data:   common.CacheRequestData{Identifier: "test", Data: value},
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error.Error())
		return
	}

	cacheResponseData := (response.Data).(common.CacheResponseData)

	// checking the saved data
	request2 := common.ThreadRequest{
		Action: common.GetAction,
		Data:   common.CacheRequestData{Identifier: cacheResponseData.Identifier},
		Nonce:  2,
	}
	in <- request2

	response2 := <-out

	cacheResponseData2 := (response2.Data).(common.CacheResponseData)

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
	request3 := common.ThreadRequest{
		Action: common.CreateAction,
		Data:   common.CacheRequestData{Identifier: cacheResponseData.Identifier, Data: value2},
		Nonce:  3,
	}
	in <- request3

	response3 := <-out

	cacheResponseData3 := (response3.Data).(common.CacheResponseData)

	if !response3.Success {
		t.Errorf("could not swap value at identifier %s\n", cacheResponseData.Identifier)
		return
	}

	if cacheResponseData.Identifier != cacheResponseData3.Identifier {
		t.Error("swapping a value should not change the identifier")
		return
	}

	request4 := common.ThreadRequest{
		Action: common.GetAction,
		Data:   common.CacheRequestData{Identifier: cacheResponseData.Identifier},
		Nonce:  4,
	}
	in <- request4

	response4 := <-out

	cacheResponseData4 := (response4.Data).(common.CacheResponseData)

	if !response4.Success {
		t.Error("could not load swapped value for verification")
	}

	if cacheResponseData4.Data != value2 {
		t.Errorf("expected swapped value to be %s but was %s\n", value2, cacheResponseData4.Data)
	}
}
