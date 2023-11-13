package scheduler

import (
	"fmt"
	"time"
)

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

func (job Job) ToString() string {

	return fmt.Sprintf("%s %s.%s (cluster: %s, config: %s)",
		job.Interval.ToString(), job.Module, job.Identifier, job.Cluster, job.Config)
}
