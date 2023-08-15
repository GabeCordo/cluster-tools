package cache

import "github.com/GabeCordo/etl/core/components/cache"

var instance *cache.Cache

func GetCacheInstance() *cache.Cache {
	if instance == nil {
		instance = cache.NewCache()
	}
	return instance
}
