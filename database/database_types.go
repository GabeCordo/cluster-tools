package database

import (
	"time"
)

const (
	MaxClusterRecordSize = 100
)

type Entry struct {
	Timestamp                     time.Time     `json:"timestamp"`
	Elapsed                       time.Duration `json:"elapsed"`
	NumProvisionedTransformRoutes int           `json:"num-provisioned-transform-routes"`
	NumProvisionedLoadRoutines    int           `json:"num-provisioned-load-routines"`
}

type Record struct {
	Entries [MaxClusterRecordSize]Entry `json:"entries"`
	Head    int8                        `json:"head"`
}

type Database struct {
	Records map[string]*Record `json:"records"`
}
