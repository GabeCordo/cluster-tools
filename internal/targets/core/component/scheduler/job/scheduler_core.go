package job

import (
	"log"
	"time"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/job"
)

// Watch
// Checks to see if a job should be added to the schedulers execution queue.
func Watch(scheduler *Scheduler) {

	// every minute we will see if the Jobs need to be statistic
	for {

		// TODO: at the moment this only works with minute scheduling

		jj := scheduler.Jobs.Get(database.Filter{})
		for _, j := range jj {
			if job.IsTimeToRun(j) {
				scheduler.mutex.Lock()
				scheduler.queue = append(scheduler.queue, *j) // todo : why are we copying?
				scheduler.mutex.Unlock()
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

// Loop
// Monitors the Job queue and executes the function if a job is found.
func Loop(scheduler *Scheduler, f func(jb job.Job) error) (err error) {

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

		// the time till the next queue check is defined in the Scheduler pipeline
		time.Sleep(time.Duration(scheduler.config.RefreshInterval) * time.Millisecond)
	}
}

func (scheduler *Scheduler) GetQueue() []job.Job {

	scheduler.mutex.RLock()
	defer scheduler.mutex.RUnlock()

	jobs := make([]job.Job, len(scheduler.queue))
	for idx, j := range scheduler.queue {
		jobs[idx] = j // make a copy
	}

	return jobs
}

func (scheduler *Scheduler) Print() {

	scheduler.Jobs.Print()
}
