package interfaces

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultPlaceholder = "*"
)

// Dump
// A static representation of the jobs in the scheduler
type Dump struct {
	Jobs []Job `yaml:"jobs"`
}

// Interval
// Contains information about how often a job should be run.
type Interval struct {
	Minute int `yaml:"minute" json:"minute" bson:"minute"` // Every Nth minute the job should run	(ex. 2 -> */2 in crontab)
	Hour   int `yaml:"hour" json:"hour" bson:"hour"`       // Every Nth hour the job should run
	Day    int `yaml:"day" json:"day" bson:"day"`          // Every Nth day the job should run
	Month  int `yaml:"month" json:"month" bson:"month"`    // Every Nth month the job should run
}

func FormatToCrontab(value int) string {

	postfix := ""
	if (value != 0) && (value != 60) {
		postfix = fmt.Sprintf("/%d", value)
	}

	return DefaultPlaceholder + postfix + " "
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

func (interval Interval) ToString() string {

	var sb strings.Builder

	sb.WriteString(FormatToCrontab(interval.Minute))
	sb.WriteString(FormatToCrontab(interval.Hour))
	sb.WriteString(FormatToCrontab(interval.Day))
	sb.WriteString(FormatToCrontab(interval.Month))

	return sb.String()
}

// Job
// Contains information about how often a module/cluster pair should
// and what config should be used during that scheduled interval.
type Job struct {
	Identifier       string            `yaml:"identifier" json:"identifier" bson:"identifier"`
	Module           string            `yaml:"module" json:"module" bson:"module"`
	Cluster          string            `yaml:"cluster" json:"cluster" bson:"cluster"`
	Config           string            `yaml:"config" json:"config" bson:"config"`
	Interval         Interval          `yaml:"interval" json:"interval" bson:"interval"`
	Metadata         map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty" bson:"metadata,omitempty"`
	lastAttemptedRun time.Time
	running          bool
}

// IsJobRunning
// returns if a job is currently marked as running.
func IsJobRunning(job *Job) bool {
	return job.running
}

// IsTimeToRun
// Returns true if a job is ready to run based on the current time the function is called.
// If the minute interval for the job is 2, we will run the job every minute that is divisible
// by two. This is similar to the '*/2' notation used by cronjob.
func IsTimeToRun(job *Job) bool {

	// we shouldn't schedule a job twice in the same period
	if job.running {
		return false
	}

	curr := time.Now().Minute()
	return (curr % job.Interval.Minute) == 0
}

func (job Job) Equals(other *Job) bool {

	if other == nil {
		return false
	}

	// the identifier is a hard comparison;
	// it doesn't matter what the other contents are, we cannot have two duplicate identifiers
	if job.Identifier == other.Identifier {
		return true
	}

	if (job.Module != other.Module) || (job.Cluster != other.Cluster) {
		return false
	}

	return job.Interval.Equals(&other.Interval)
}

func (job Job) ToString() string {

	return fmt.Sprintf("%s %s.%s (cluster: %s, config: %s)",
		job.Interval.ToString(), job.Module, job.Identifier, job.Cluster, job.Config)
}

// Filter
// A struct containing the ways one can filter out jobs upon search
type Filter struct {
	Identifier string
	Module     string
	Cluster    string
	Interval   Interval
}

func (filter Filter) UseIdentifier() bool {
	return filter.Identifier != ""
}

func (filter Filter) UseModule() bool {
	return filter.Interval.Empty() && (filter.Module != "") && (filter.Cluster == "")
}

func (filter Filter) UseCluster() bool {
	return filter.Interval.Empty() && (filter.Module != "") && (filter.Cluster != "")
}

func (filter Filter) UseInterval() bool {

	return !filter.Interval.Empty() && (filter.Module != "") && (filter.Cluster != "")
}
