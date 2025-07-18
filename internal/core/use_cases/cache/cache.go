package cache

import (
	"errors"
	"github.com/GabeCordo/Flock/internal/core/component/cache"
)

type UseCase struct {
	CacheComponent cache.Cache
}

func (uc UseCase) Save(identifier string, data any, expiresIn float64) (newIdentifier string, err error) {

	err = nil

	if _, found := uc.CacheComponent.Get(identifier); found {
		if ok := uc.CacheComponent.Swap(identifier, data, expiresIn); !ok {
			err = errors.New("failed to swap CacheComponent")
		}
		newIdentifier = identifier
	} else {
		// what if the user forgets to pass in an expiry time that's now set to 0?
		if expiresIn == 0 {
			newIdentifier = uc.CacheComponent.Save(data)
		} else {
			newIdentifier = uc.CacheComponent.Save(data, expiresIn)
		}
	}

	return newIdentifier, err
}

func (uc UseCase) Load(identifier string) (data any, err error) {

	cacheData, isFoundAndNotExpired := uc.CacheComponent.Get(identifier)
	if !isFoundAndNotExpired {
		err = errors.New("no data can be found with that identifier because it does not exist or has expired")
	} else {
		data = cacheData
	}

	return data, err
}
