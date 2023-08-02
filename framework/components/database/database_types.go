package database

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"sync"
	"time"
)

const (
	MaxClusterRecordSize = 124
	Empty                = -1
)

type DataType uint8

const (
	Statistic DataType = iota
	Config
	Module
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
	Records map[string]map[string]*Record        `json:"record"`
	Configs map[string]map[string]cluster.Config `json:"configs"`

	mutex sync.RWMutex
}
