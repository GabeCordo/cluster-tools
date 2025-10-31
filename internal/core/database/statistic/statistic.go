package statistic

import (
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/plover"
	"time"
)

type Statistic struct {
	Timestamp time.Time          `json:"timestamp"`
	Namespace string             `json:"namespace"`
	Pipeline  string             `json:"pipeline"`
	Elapsed   time.Duration      `json:"elapsed"`
	Data      *plover.Statistics `json:"statistics"`
}

type Database interface {
	Get(filter database.Filter) []*plover.Statistics
	Create(filter database.Filter, record *plover.Statistics) (*plover.Statistics, error)
	Replace(filter database.Filter, record *plover.Statistics) error
	Delete(filter database.Filter) error
	Save(path string) error
	Load(path string) error
	Print()
}
