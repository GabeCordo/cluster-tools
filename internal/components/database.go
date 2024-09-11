package components

import (
	"github.com/GabeCordo/cluster-tools/internal/interfaces"
	"time"
)

type ConfigFilter struct {
	Module     string
	Identifier string
}

type Database interface {
	Save(path string) error
	Load(path string) error
	Print()
}

type ConfigDatabase interface {
	Get(filter ConfigFilter) (records []interfaces.Config, err error)
	Create(moduleIdentifier, configIdentifier string, cfg interfaces.Config) (err error)
	Replace(moduleIdentifier, configIdentifier string, cfg interfaces.Config) (err error)
	Delete(moduleIdentifier, configIdentifier string) (err error)
}

type JobDatabase interface {
	GetAll() ([]interfaces.Job, error)
	GetBy(filter *interfaces.Filter) ([]interfaces.Job, error)
	Create(job *interfaces.Job) error
	Delete(filter *interfaces.Filter) error
}

type Statistic struct {
	Timestamp time.Time             `json:"timestamp"`
	Elapsed   time.Duration         `json:"elapsed"`
	Stats     interfaces.Statistics `json:"statistics"`
}

type StatisticFilter struct {
	Module  string
	Cluster string
	Verbose bool
}

type StatisticDatabase interface {
	Get(filter StatisticFilter) (records []Statistic, err error)
	Create(moduleId, clusterId string, statistic Statistic) (err error)
	Delete(moduleId string) (err error)
}
