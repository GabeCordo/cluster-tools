package cache

import "github.com/GabeCordo/mango-core/core/components/cache"

var instance *cache.Cache

func GetCacheInstance() *cache.Cache {
	if instance == nil {
		instance = cache.NewCache()
	}
	return instance
}
