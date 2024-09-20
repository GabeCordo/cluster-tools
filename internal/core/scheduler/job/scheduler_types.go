package job

import (
	"github.com/GabeCordo/cluster-tools/internal/database"
	"github.com/GabeCordo/cluster-tools/internal/database/job"
	"sync"
)

// Config
// Contains dynamic runtime information use by the Scheduler on
// startup of the program.
type Config struct {
	RefreshInterval int `yaml:"refresh_interval"` // how often the scheduler will check to see if new Jobs should be added to the run queue.
}

// Scheduler
// Contains a collection of Jobs that are run on fixed intervals.
type Scheduler struct {
	Jobs   database.Database // A static list of Jobs registered to the scheduler.
	queue  []job.Job         // A dynamic list of Jobs waiting to be run.
	config Config            // Dynamic information that tells the Scheduler how to run.
	mutex  sync.RWMutex
}

// New
// Creates a new scheduler and initializes default fields.
func New(database database.Database) (*Scheduler, error) {
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
