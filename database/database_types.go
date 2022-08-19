package database

import (
	"ETLFramework/cluster"
	"sync"
	"time"
)

const (
	MaxClusterRecordSize = 124
	Empty                = -1
)

type RecordType uint8

const (
	Statistics RecordType = 0
	Monitors              = 1
)

type Record interface {
	Empty() bool
}

type Entry struct {
	Timestamp time.Time          `json:"timestamp"`
	Elapsed   time.Duration      `json:"elapsed"`
	Stats     cluster.Statistics `json:"statistics"`
}

type StatisticRecord struct {
	Entries [MaxClusterRecordSize]Entry `json:"entries"` // IMMUTABLE
	Head    int8                        `json:"head"`

	mutex sync.Mutex
}

type MonitorId struct {
	Identifier string
	Id         int8
}

type Monitor struct {
	Id                   int
	NumExtractRoutines   int
	NumTransformRoutines int
	NumLoadRoutines      int
}

type MonitorRecord struct {
	Entries [cluster.MaxConcurrentMonitors]Monitor `json:"monitors"` // MUTABLE
	Head    int8

	mutex sync.Mutex
}

type Database struct {
	statistics map[string]*StatisticRecord `json:"statistics"`
	monitors   map[string]*MonitorRecord   `json:"monitors"`

	mutex sync.Mutex
}
