package database

import (
	"fmt"
	"strings"
)

// Interval
// Contains information about how often a job should be statistic.
type Interval struct {
	Minute int `yaml:"minute" json:"minute" bson:"minute"` // Every Nth minute the job should statistic	(ex. 2 -> */2 in crontab)
	Hour   int `yaml:"hour" json:"hour" bson:"hour"`       // Every Nth hour the job should statistic
	Day    int `yaml:"day" json:"day" bson:"day"`          // Every Nth day the job should statistic
	Month  int `yaml:"month" json:"month" bson:"month"`    // Every Nth month the job should statistic
}

func (interval Interval) Empty() bool {

	return (interval.Month == 0) && (interval.Day == 0) && (interval.Hour == 0) && (interval.Minute == 0)
}

func (interval Interval) Equals(other *Interval) bool {

	if other == nil {
		return false
	}

	return interval.Hour == other.Hour &&
		interval.Day == other.Day &&
		interval.Month == other.Month &&
		interval.Minute == other.Minute
}

func (interval Interval) formatToCron(value int) string {

	postfix := ""
	if (value != 0) && (value != 60) {
		postfix = fmt.Sprintf("/%d", value)
	}

	return "*" + postfix + " "
}

func (interval Interval) ToString() string {

	var sb strings.Builder

	sb.WriteString(interval.formatToCron(interval.Minute))
	sb.WriteString(interval.formatToCron(interval.Hour))
	sb.WriteString(interval.formatToCron(interval.Day))
	sb.WriteString(interval.formatToCron(interval.Month))

	return sb.String()
}

const Empty = ""

type Filter struct {
	Namespace  string
	Pipeline   string
	Interval   Interval
	Processor  uint64
	Identifier string
	Config     string
	Verbose    bool
}

func (filter Filter) IsEmpty() bool {
	return filter.Namespace == "" && filter.Pipeline == "" && filter.Identifier == "" && filter.Verbose == false
}

func (filter Filter) UseIdentifier() bool {
	return filter.Identifier != ""
}

func (filter Filter) UseNamespace() bool {
	return (filter.Namespace != "") && (filter.Pipeline == "")
}

func (filter Filter) UsePipeline() bool {
	return (filter.Namespace != "") && (filter.Pipeline != "") && filter.Interval.Empty()
}

func (filter Filter) UseInterval() bool {
	return !filter.Interval.Empty() && (filter.Namespace != "") && (filter.Pipeline != "")
}

type Database interface {
	Get(filter Filter) []any
	Create(filter Filter, record any) (any, error)
	Replace(filter Filter, record any) error
	Delete(filter Filter) error
	Save(path string) error
	Load(path string) error
	Print()
}
