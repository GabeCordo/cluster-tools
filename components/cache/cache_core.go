package cache

import (
	"github.com/GabeCordo/fack"
	"time"
)

func (record Record) IsExpired() bool {
	return time.Now().Sub(record.created).Minutes() > record.expiry
}

func (cache *Cache) Save(data any, expiry ...float64) string {

	var record Record
	if len(expiry) == 1 {
		record = Record{data, time.Now(), expiry[0]}
	} else {
		record = Record{data, time.Now(), DefaultCacheExpiry}
	}

	var identifier string
	for {
		identifier = fack.GenerateRandomString(6)
		if _, found := cache.records.Load(identifier); found {
			continue
		} else {
			break
		}
	}

	cache.records.Store(identifier, record)
	return identifier
}

func (cache *Cache) Swap(identifier string, data any, expiry ...float64) bool {

	if value, found := cache.records.Load(identifier); found {
		record := (value).(Record)

		record.data = data
		record.created = time.Now()

		cache.records.Store(identifier, record)
		return true
	} else {
		return false
	}
}

func (cache *Cache) Get(identifier string) (any, bool) {
	if data, found := cache.records.Load(identifier); found && !(data).(Record).IsExpired() {
		return (data).(Record), true
	} else {
		return nil, false
	}
}

func (cache *Cache) Remove(identifier string) {

	if _, found := cache.records.Load(identifier); found {
		cache.records.Delete(identifier)
	}
}

func (cache *Cache) Clean() {

	cache.records.Range(func(key any, value any) bool {
		identifier := (key).(string)
		record := (value).(Record)

		if record.IsExpired() {
			cache.records.Delete(identifier)
		}

		return false // stop iteration
	})
}
