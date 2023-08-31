package api

import "testing"

var clientHost = "http://localhost:8136"

var key = "test"
var value = "this is a test"

func TestCache(t *testing.T) {

	_, err := Cache(host, key, value)
	if err != nil {
		t.Error(err)
	}
}

func TestGetFromCache(t *testing.T) {

	random, err := Cache(host, key, value)
	if err != nil {
		t.Error(err)
	}

	pulledValue, err := GetFromCache(host, random)
	if err != nil {
		t.Error(err)
	}

	if pulledValue != value {
		t.Error("pulled cache does not match expected cached value")
	}
}
