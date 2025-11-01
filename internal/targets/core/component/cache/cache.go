package cache

import "errors"

var TooManyRecords = errors.New("the number of records inside the cache exceeds the maximum number of records allowed")

var IdentifierAlreadyExists = errors.New("attempted to save a record inside the cache with an existing identifier")

const CreateIdentifier = ""

type Cache interface {
	Save(identifier string, data any, expiry ...float64) (string, error)
	Swap(identifier string, data any, expiry ...float64) bool
	Get(identifier string) (any, bool)
	Remove(identifier string)
	Clean()
}
