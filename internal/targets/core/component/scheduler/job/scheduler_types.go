package job

import (
	"sync"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

// Config
// Contains dynamic runtime information use by the Scheduler on
// startup of the program.
type Config struct {
	RefreshInterval int `yaml:"refresh_interval"` // how often the scheduler will check to see if new Jobs should be added to the statistic queue.
}

// Scheduler
// Contains a collection of Jobs that are statistic on fixed intervals.
type Scheduler struct {
	Jobs   job.Database // A static list of Jobs registered to the scheduler.
	queue  []job.Job    // A dynamic list of Jobs waiting to be statistic.
	config Config       // Dynamic information that tells the Scheduler how to statistic.
	mutex  sync.RWMutex
}

// New
// Creates a new scheduler and initializes default fields.
func New(database job.Database) (*Scheduler, error) {
	scheduler := new(Scheduler)
	scheduler.Jobs = database
	scheduler.queue = make([]job.Job, 0)
	scheduler.config.RefreshInterval = 1 // ms
	return scheduler, nil
}

func (scheduler *Scheduler) ItemsInQueue() int {
	scheduler.mutex.RLock()
	defer scheduler.mutex.RUnlock()

	return len(scheduler.queue)
}
