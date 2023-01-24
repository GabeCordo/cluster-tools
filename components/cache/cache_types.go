package cache

import (
	"sync"
	"time"
)

const (
	DefaultCacheExpiry       = 2.0   // minutes
	DefaultMaxAllowedRecords = 0x3E8 // the default maximum is 1000
)

type Record struct {
	data    any
	created time.Time
	expiry  float64
}

type Cache struct {
	records sync.Map

	maxAllowedRecords uint32
	numOfRecords      uint32
}

func NewCache(maxAllowedRecords ...uint32) *Cache {
	cache := new(Cache)
	if cache == nil {
		panic("system ran out of memory")
	}
	if len(maxAllowedRecords) == 1 {
		cache.maxAllowedRecords = maxAllowedRecords[0]
	} else {
		cache.maxAllowedRecords = DefaultMaxAllowedRecords
	}
	cache.numOfRecords = 0

	return cache
}
