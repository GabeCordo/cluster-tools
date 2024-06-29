package scheduler

import (
	"github.com/GabeCordo/cluster-tools/internal/core/components/database"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"log"
	"time"
)

// Watch
// Checks to see if a job should be added to the schedulers execution queue.
func Watch(scheduler *Scheduler) {

	// every minute we will see if the Jobs need to be run
	for {

		// TODO: at the moment this only works with minute scheduling

		jobs, _ := scheduler.Jobs.GetAll()
		for idx, job := range jobs {

			if interfaces.IsTimeToRun(&job) {
				scheduler.mutex.Lock()
				scheduler.queue = append(scheduler.queue, &jobs[idx])
				scheduler.mutex.Unlock()
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

// Loop
// Monitors the Job queue and executes the function if a job is found.
func Loop(scheduler *Scheduler, f func(job *interfaces.Job) error) (err error) {

	// loop over the job queue until one of the elements hits an error
	for {

		scheduler.mutex.RLock()
		if len(scheduler.queue) >= 1 {
			// pop the first element of the queue (FIFO) and remove the
			// first element by slicing out the first element
			popped := scheduler.queue[0]
			scheduler.queue = scheduler.queue[1:]

			// outdated;
			// to abide by the pattern, if the called function returns an
			// error stop the scheduler loop.
			//
			// note: stopping the scheduler silently can create problems
			// 		 during long runtimes. How does the operator know when
			//		 the scheduler no longer operates? it doesn't.
			if err = f(popped); err != nil {
				log.Println(err)
				// outdated:
				// if we receive a non-nil code, an error has occurred, so re-append
				// the popped job to the back of the queue to try again later
				//if err != nil {
				//	scheduler.queue = append(scheduler.queue, popped)
				//}
			}
		}
		scheduler.mutex.RUnlock()

		// the time till the next queue check is defined in the Scheduler config
		time.Sleep(time.Duration(scheduler.config.RefreshInterval) * time.Millisecond)
	}

	return err
}

func (scheduler *Scheduler) GetQueue() []interfaces.Job {

	scheduler.mutex.RLock()
	defer scheduler.mutex.RUnlock()

	jobs := make([]interfaces.Job, len(scheduler.queue))
	for idx, job := range scheduler.queue {
		jobs[idx] = *job // make a copy
	}

	return jobs
}

func (scheduler *Scheduler) Print() {

	if db, ok := (scheduler.Jobs).(database.Database); ok {
		db.Print()
	}
}
