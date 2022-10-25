package database

import (
	"etl/components/cluster"
	"sync"
	"time"
)

const (
	MaxClusterRecordSize = 124
	Empty                = -1
)

type Entry struct {
	Timestamp time.Time          `json:"timestamp"`
	Elapsed   time.Duration      `json:"elapsed"`
	Stats     cluster.Statistics `json:"statistics"`
}

type Record struct {
	Entries [MaxClusterRecordSize]Entry `json:"entries"` // IMMUTABLE
	Head    int8                        `json:"head"`

	mutex sync.Mutex
}

type Database struct {
	Records map[string]*Record `json:"record"`

	mutex sync.Mutex
}
