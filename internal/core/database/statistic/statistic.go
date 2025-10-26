package statistic

import (
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
