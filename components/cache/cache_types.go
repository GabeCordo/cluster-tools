package cache

import (
	"sync"
	"time"
)

const (
	DefaultCacheExpiry = 2.0 // minutes
)

type Record struct {
	data    any
	created time.Time
	expiry  float64
}

type Cache struct {
	records sync.Map
}

func NewCache() *Cache {
	cache := new(Cache)
	if cache == nil {
		panic("system ran out of memory")
	}
	return cache
}
