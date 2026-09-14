package statistic

import (
	"github.com/GabeCordo/ScalingFunctions"
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
)

type Statistic struct {
	Timestamp time.Time                    `json:"timestamp"`
	Namespace string                       `json:"namespace"`
	Pipeline  string                       `json:"pipeline"`
	Elapsed   time.Duration                `json:"elapsed,omitempty"`
	Data      *ScalingFunctions.Statistics `json:"statistics"`
}

type Database interface {
	Get(filter database.Filter) []*ScalingFunctions.Statistics
	Create(filter database.Filter, record *ScalingFunctions.Statistics) (*ScalingFunctions.Statistics, error)
	Replace(filter database.Filter, record *ScalingFunctions.Statistics) error
	Delete(filter database.Filter) error
	Distinct(filter database.Filter) ([]any, error)
	Save(path string) error
	Load(path string) error
	Print()
}
