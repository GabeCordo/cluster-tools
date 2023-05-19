package database

import (
	"github.com/GabeCordo/etl/components/cluster"
	"sync"
	"time"
)

const (
	MaxClusterRecordSize = 124
	Empty                = -1
)

type DataType uint8

const (
	Statistic DataType = 0
	Config             = 1
)

type Entry struct {
	Timestamp time.Time          `json:"timestamp"`
	Elapsed   time.Duration      `json:"elapsed"`
	Stats     cluster.Statistics `json:"statistics"`
}

type Record struct {
	Entries [MaxClusterRecordSize]Entry `json:"entries"` // IMMUTABLE
	Head    int8                        `json:"head"`

	mutex sync.RWMutex
}

type Database struct {
	Records map[string]*Record        `json:"record"`
	Configs map[string]cluster.Config `json:"configs"`

	mutex sync.RWMutex
}
