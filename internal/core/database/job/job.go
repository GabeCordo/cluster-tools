package job

import (
	"fmt"
	"time"

	"github.com/GabeCordo/Flock/internal/core/database"
)

// Dump
// A static representation of the jobs in the scheduler
type Dump struct {
	Jobs []Job `yaml:"jobs"`
}

// Job
// Contains information about how often a module/cluster pair should
// and what pipeline should be used during that scheduled interval.
type Job struct {
	Identifier       string            `yaml:"identifier" json:"identifier" bson:"identifier"`
	Namespace        string            `yaml:"namespace" json:"namespace" bson:"namespace"`
	Pipeline         string            `yaml:"pipeline" json:"pipeline" bson:"pipeline"`
	Interval         database.Interval `yaml:"interval" json:"interval" bson:"interval"`
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
// Returns true if a job is ready to statistic based on the current time the function is called.
// If the minute interval for the job is 2, we will statistic the job every minute that is divisible
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

	if (job.Namespace != other.Namespace) || (job.Pipeline != other.Pipeline) {
		return false
	}

	return job.Interval.Equals(&other.Interval)
}

func (job Job) ToString() string {

	return fmt.Sprintf("%s %s.%s (pipeline: %s)",
		job.Interval.ToString(), job.Namespace, job.Identifier, job.Pipeline)
}
