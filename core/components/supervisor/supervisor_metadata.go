package supervisor

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"github.com/GabeCordo/fack"
	"sync"
)

type M struct {
	data map[string]string

	mutex sync.RWMutex
}

func NewMetadata(data map[string]string) cluster.M {

	metadata := new(M)

	if data != nil {
		metadata.data = data
	} else {
		metadata.data = make(map[string]string)
	}

	return metadata
}

func (metadata *M) GetKey(key string) string {

	metadata.mutex.RLock()
	defer metadata.mutex.RUnlock()

	if value, found := metadata.data[key]; found {
		return value
	} else {
		return fack.EmptyString
	}
}
