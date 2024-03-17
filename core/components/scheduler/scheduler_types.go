package scheduler

import (
	"sync"
	"time"
)

// Interval
// Contains information about how often a job should be run.
type Interval struct {
	Minute int `yaml:"minute"` // Every Nth minute the job should run	(ex. 2 -> */2 in crontab)
	Hour   int `yaml:"hour"`   // Every Nth hour the job should run
	Day    int `yaml:"day"`    // Every Nth day the job should run
	Month  int `yaml:"month"`  // Every Nth month the job should run
}

// Job
// Contains information about how often a module/cluster pair should
// and what config should be used during that scheduled interval.
type Job struct {
	Identifier       string            `yaml:"identifier"`
	Module           string            `yaml:"module"`
	Cluster          string            `yaml:"cluster"`
	Config           string            `yaml:"config"`
	Interval         Interval          `yaml:"interval"`
	Metadata         map[string]string `yaml:"metadata"`
	lastAttemptedRun time.Time         `yaml:"lastAttemptedRun"`
	running          bool
}

// Dump
// A static representation of the jobs in the scheduler
type Dump struct {
	Config Config `yaml:"config"`
	Jobs   []Job  `yaml:"jobs"`
}

// Filter
// A struct containing the ways one can filter out jobs upon search
type Filter struct {
	Identifier string
	Module     string
	Cluster    string
	Interval   Interval
}

// Config
// Contains dynamic runtime information use by the Scheduler on
// startup of the program.
type Config struct {
	RefreshInterval int `yaml:"refresh_interval"` // how often the scheduler will check to see if new jobs should be added to the run queue.
}

// Scheduler
// Contains a collection of jobs that are run on fixed intervals.
type Scheduler struct {
	jobs   []Job  // A static list of jobs registered to the scheduler.
	queue  []*Job // A dynamic list of jobs waiting to be run.
	config Config // Dynamic information that tells the Scheduler how to run.
	mutex  sync.RWMutex
}

// New
// Creates a new scheduler and initializes default fields.
func New() (*Scheduler, error) {

	scheduler := new(Scheduler)
	scheduler.jobs = make([]Job, 0)
	scheduler.queue = make([]*Job, 0)
	scheduler.config.RefreshInterval = 1 // ms
	return scheduler, nil
}

func (scheduler *Scheduler) ItemsInQueue() int {
	scheduler.mutex.RLock()
	defer scheduler.mutex.RUnlock()

	return len(scheduler.queue)
}
