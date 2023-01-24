package core

var provisionerMemory *ProvisionerMemory

func GetProvisionerMemoryInstance() *ProvisionerMemory {
	if provisionerMemory == nil {
		provisionerMemory = NewProvisionerResponses()
	}

	return provisionerMemory
}

func (provisionerResponses ProvisionerMemory) DoesCacheResponseExists(nonce uint32) bool {
	_, found := provisionerResponses.cache.Load(nonce)
	return found
}

func (provisionerResponses ProvisionerMemory) PopCacheResponse(nonce uint32) (CacheResponse, bool) {
	response, found := provisionerResponses.cache.Load(nonce)
	if found {
		provisionerResponses.cache.Delete(nonce)
	}

	if response == nil {
		return CacheResponse{}, found
	} else {
		return (response).(CacheResponse), found
	}
}

func (provisionerResponses ProvisionerMemory) DoesDatabaseResponseExists(nonce uint32) bool {
	_, found := provisionerResponses.database.Load(nonce)
	return found
}

func (provisionerResponses ProvisionerMemory) PopDatabasedResponse(nonce uint32) (DatabaseResponse, bool) {
	response, found := provisionerResponses.database.Load(nonce)
	if found {
		provisionerResponses.database.Delete(nonce)
	}

	if response == nil {
		return DatabaseResponse{}, found
	} else {
		return (response).(DatabaseResponse), found
	}
}
