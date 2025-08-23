package local

import (
	"bytes"
	"fmt"
	cache2 "github.com/FortifiedCode/flock/internal/core/component/cache"
	"log"
	"math/rand"
	"time"
)

const (
	DefaultCacheRecordIdentifierSize = 15
)

// integer constants
const (
	maxGeneratedStringLength = 32
	lowerASCIIBound          = 97
	upperASCIIBound          = 122
)

func RandInteger(min int, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404 -- Rand does not need to be cryptographically secure
	return min + r.Intn(max-min)
}

func GenerateRandomString() string {
	buffer := new(bytes.Buffer)
	for i := 0; i < maxGeneratedStringLength; i++ {
		char := RandInteger(lowerASCIIBound, upperASCIIBound)
		buffer.WriteString(fmt.Sprint(char))
	}
	return buffer.String()
}

func (cache *Cache) Save(identifier string, data any, expiry ...float64) (string, error) {
	cache.m.Lock()
	defer cache.m.Unlock()

	// if the system is being statistic on a low-memory machine, it
	// is important that the cache does not grow too large and
	// take away resources from the os or other etl processes.
	if cache.numOfRecords == cache.maxAllowedRecords {
		// send a warning to the developer that notifies them that they are likely
		// abusing the cache, it is a short-term data storage for inter-cluster
		// communication. If they just have too many clusters using the cache, it
		// might indicate that they need to increase the ram on their production environment
		log.Println("(warning) cache miss, increase the maximum number of records allowed.")
		log.Println("[ increasing the maximum records on low-ram machines will degrade performance, be careful ]")
		return "", cache2.TooManyRecords
	}

	var record Record
	if len(expiry) == 1 {
		record = Record{data, time.Now(), expiry[0]}
	} else {
		record = Record{data, time.Now(), DefaultCacheExpiry}
	}

	isIdentifierProvided := identifier != cache2.CreateIdentifier

	// the user has the option to choose their own identifier, before we can accept
	// this value, verify that the identifier does not already exist in the cache.
	//
	// when the value already exists in the cache, we should fail.
	_, found := cache.records.Load(identifier)
	if isIdentifierProvided && found {
		// let the user use the Swap function if they wish to override the value in the cache
		return "", cache2.IdentifierAlreadyExists
	}

	if !isIdentifierProvided {
		// generate a unique string until there are no conflicts in the cache.
		for {
			identifier = GenerateRandomString()

			// in the odd case the cache identifier already exists, try again until we find a unique id
			// Note: this should not hit as records (should) consistently be deleted
			if _, found := cache.records.Load(identifier); !found {
				break
			}
		}
	}

	cache.records.Store(identifier, record)
	// the numOfRecords will be used to track whether the cache reaches its maximum size
	cache.numOfRecords++

	return identifier, nil
}

func (cache *Cache) Swap(identifier string, data any, expiry ...float64) bool {
	cache.m.Lock()
	defer cache.m.Unlock()

	if value, found := cache.records.Load(identifier); found {
		record, ok := (value).(Record)
		if !ok {
			return false
		}

		record.data = data
		record.created = time.Now()

		// if we are provided with a new expiry time, use that, else re-use the old one
		// TODO : I don't really see a use for this? - I could be wrong in the future
		//if len(expiry) == 1 {
		//	record.expiry = expiry[0]
		//}

		cache.records.Swap(identifier, record)
		return true
	} else {
		// no record exists with that identifier, there's nothing to "swap"
		return false
	}
}

// Get
// todo ~ lot's of type casting could be optimized
func (cache *Cache) Get(identifier string) (any, bool) {
	cache.m.RLock()
	defer cache.m.RUnlock()

	// requirements for a get operation:
	// 1) identifier exists
	// 2) record expiry has not been hit
	if data, found := cache.records.Load(identifier); found && !(data).(Record).IsExpired() {
		return (data).(Record).data, true // valid
	} else {
		return nil, false // not found or expired
	}

	// no affect to the record count
}

func (cache *Cache) Remove(identifier string) {
	cache.m.Lock()
	defer cache.m.Unlock()

	if _, found := cache.records.Load(identifier); found {
		cache.records.Delete(identifier)
		// one less record in the cache, "release" that space for another record
		cache.numOfRecords--
	}
}

// Clean
// This function is left upto the developer to invoke for performance reasons.
// Allows Developers to remove all expired records in one "sweep" to avoid
// hitting expired records when calling the Cache.Get() function.
func (cache *Cache) Clean() {
	cache.m.Lock()
	defer cache.m.Unlock()

	cache.records.Range(func(key any, value any) bool {
		identifier := (key).(string)
		record := (value).(Record)

		if record.IsExpired() {
			cache.records.Delete(identifier)

			// one less record in the cache, "release" that space for another record
			cache.numOfRecords--
		}

		return false // returning false stops iteration in the Range function
	})
}
