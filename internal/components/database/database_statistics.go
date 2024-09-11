package database

import (
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"time"
)

type Statistic struct {
	Timestamp time.Time             `json:"timestamp"`
	Elapsed   time.Duration         `json:"elapsed"`
	Stats     interfaces.Statistics `json:"statistics"`
}

type StatisticDatabase interface {
	Get(filter StatisticFilter) (records []Statistic, err error)
	Create(moduleId, clusterId string, statistic Statistic) (err error)
	Delete(moduleId string) (err error)
}
