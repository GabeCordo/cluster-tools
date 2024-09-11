package database

import "errors"

var StoreTypeMismatch = errors.New("the received type and desired store type do not match")
