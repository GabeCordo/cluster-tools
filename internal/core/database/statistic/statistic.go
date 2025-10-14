package statistic

import (
	"github.com/FortifiedCode/plover"
	"time"
)

type Wrapper struct {
	Timestamp time.Time         `json:"timestamp"`
	Namespace string            `json:"namespace"`
	Pipeline  string            `json:"pipeline"`
	Elapsed   time.Duration     `json:"elapsed"`
	Stats     plover.Statistics `json:"statistics"`
}
