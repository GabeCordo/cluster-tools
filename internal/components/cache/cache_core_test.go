package cache

import (
	"testing"
	"time"
)

func TestCache_Save(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	cache.Save("test")

	if cache.numOfRecords != 1 {
		t.Error("cache record counter not working")
	}
}

// TestCache_Save
func TestCache_SaveNoExpiryParam(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	identifier := cache.Save("test")

	value, found := cache.records.Load(identifier)
	if !found {
		t.Errorf("expected to find value with identifier %s\n", identifier)
		return
	}

	record, castOk := (value).(Record)
	if !castOk {
		t.Error("expected the stored value to be of type Record")
		return
	}

	if record.expiry != DefaultCacheExpiry {
		t.Error("expected the expiry to be the DefaultCacheExpiry if no parameter was provided")
	}
}

func TestCache_SaveExpiryParam(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	expiresInNSec := 4.0
	identifier := cache.Save("test", expiresInNSec)

	value, found := cache.Get(identifier)
	if !found {
		t.Errorf("expected to find value with identifier %s\n", identifier)
		return
	}

	_, castOk := (value).(string)
	if !castOk {
		t.Error("expected the stored value to be of type Record")
		return
	}
}

func TestCache_SaveMaxRecordsReached(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}
	cache.numOfRecords = DefaultMaxAllowedRecords

	if identifier := cache.Save("test"); identifier != "" {
		t.Error("expected no value to be saved if the max records is reached")
	}
}

// TestCache_Get
// The cache.get function should return the value and true if the identifier exists
func TestCache_Get(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	value := "test"
	identifier := cache.Save(value)

	if foundValue, found := cache.Get(identifier); !found {
		t.Error("cache did not return value that exists")
	} else if foundValue != value {
		t.Errorf("expected cache to get value %s but found %s\n", value, foundValue)
	}
}

// TestCache_Get2
// The cache.get function shall return false if no record with that identifier exists
func TestCache_Get2(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	if _, found := cache.Get("test"); found {
		t.Error("cache.get returned true when no value should exist")
	}
}

// TestCache_Get3
// The cache.get function shall return false if a record exists but is expired
// Note: this could result in a memory leak if not cleaned up
func TestCache_Get3(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	identifier := "test"
	fourMinAgo := time.Now().Add(-1 * 4 * time.Minute)
	record := Record{created: fourMinAgo, expiry: DefaultCacheExpiry}
	cache.records.Store(identifier, record)

	if _, found := cache.Get(identifier); found {
		t.Error("cache should mark identifier as not found if expired")
	}
}

// TestCache_Remove
// The cache.remove function shall remove the value from the internal map and
// decrement the numOfRecords pointer
func TestCache_Remove(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	identifier := cache.Save("foo")

	if cache.numOfRecords != 1 {
		t.Error("expected the cache to have 1 record saved")
		return
	}

	cache.Remove(identifier)

	if cache.numOfRecords != 0 {
		t.Error("cache not decrementing the numOfRecord counter after delete")
	}
}

// TestCache_Remove2
// The cache numOfRecords counter shall remain the same if a caller attempts to delete
// a non-existent record from the cache
func TestCache_Remove2(t *testing.T) {

	cache := NewCache(DefaultMaxAllowedRecords)
	if cache == nil {
		t.Error("cache pointer should not be nil")
		return
	}

	initialNumOfRecordsCounter := cache.numOfRecords

	// remove a non-existent identifier
	cache.Remove("foo")

	if initialNumOfRecordsCounter != cache.numOfRecords {
		t.Error("the cache numOfRecords counter changed when it shouldn't have")
	}
}
