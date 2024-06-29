package scheduler

import (
	"github.com/GabeCordo/cluster-tools/internal/core/components/database"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
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
	Jobs   database.JobDatabase // A static list of Jobs registered to the scheduler.
	queue  []*interfaces.Job    // A dynamic list of Jobs waiting to be run.
	config Config               // Dynamic information that tells the Scheduler how to run.
	mutex  sync.RWMutex
}

// New
// Creates a new scheduler and initializes default fields.
func New(database database.JobDatabase) (*Scheduler, error) {
	scheduler := new(Scheduler)
	scheduler.Jobs = database
	scheduler.queue = make([]*interfaces.Job, 0)
	scheduler.config.RefreshInterval = 1 // ms
	return scheduler, nil
}

func (scheduler *Scheduler) ItemsInQueue() int {
	scheduler.mutex.RLock()
	defer scheduler.mutex.RUnlock()

	return len(scheduler.queue)
}
