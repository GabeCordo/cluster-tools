package scheduler

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Load
// Loads a set of static jobs defined in a yaml file into runtime.
func Load(scheduler *Scheduler, path string) error {

	// if the path does not exist, or the path does and is not a directory, stop
	// we are looking for a folder that has yaml files with jobs
	if fInfo, err := os.Stat(path); os.IsNotExist(err) || (os.IsExist(err) && !fInfo.IsDir()) {
		output := fmt.Sprintf("%s is not a valid directory on the system", path)
		return errors.New(output)
	}

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {

		// we don't care to open files that are directories
		if d.IsDir() {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		dump := &Dump{}
		if err = yaml.Unmarshal(b, dump); err != nil {
			return err
		}

		if err = Initialize(scheduler, dump); err != nil {
			return err
		}

		return nil
	})

	// if filepath.walkdir returns an error, it will be passed here
	return err
}

// Initialize
// Copy the contents of a scheduler dump into the scheduler memory.
func Initialize(scheduler *Scheduler, dump *Dump) error {

	scheduler.config = dump.Config

	for _, job := range dump.Jobs {
		scheduler.jobs = append(scheduler.jobs, job)
	}

	return nil
}

// Save
// Move all jobs in the scheduler into files seperated by their modules.
func Save(scheduler *Scheduler, path string) error {

	if fInfo, err := os.Stat(path); os.IsNotExist(err) || (os.IsExist(err) && !fInfo.IsDir()) {
		output := fmt.Sprintf("%s is not a valid path to a directory", path)
		return errors.New(output)
	}

	// clear all the files that already existed in the folder, they should have been loaded
	// into the scheduler if the Load function was called correctly
	os.RemoveAll(path)
	os.MkdirAll(path, 0750)

	moduleSeperatedJobs := make(map[string][]Job)

	// order all the jobs by the module they belong to
	for _, job := range scheduler.jobs {
		if _, found := moduleSeperatedJobs[job.Module]; !found {
			moduleSeperatedJobs[job.Module] = make([]Job, 0)
		}
		moduleSeperatedJobs[job.Module] = append(moduleSeperatedJobs[job.Module], job)
	}

	// save each module's job into its own file
	for module, jobs := range moduleSeperatedJobs {

		filePath := fmt.Sprintf("%s/schedule_%s.yml", path, module)
		dump := &Dump{Config: scheduler.config, Jobs: jobs}
		b, err := yaml.Marshal(dump)
		if err != nil {
			output := fmt.Sprintf("failed to turn jobs into dump file %s", err.Error())
			return errors.New(output)
		}
		if err = os.WriteFile(filePath, b, 0750); err != nil {
			output := fmt.Sprintf("failed to write dump to file %s", err.Error())
			return errors.New(output)
		}
	}

	return nil
}

// Watch
// Checks to see if a job should be added to the schedulers execution queue.
func Watch(scheduler *Scheduler) {

	// every minute we will see if the jobs need to be run
	for {

		// TODO: at the moment this only works with minute scheduling

		for idx, job := range scheduler.jobs {

			if IsTimeToRun(&job) {
				scheduler.mutex.Lock()
				scheduler.queue = append(scheduler.queue, &scheduler.jobs[idx])
				scheduler.mutex.Unlock()
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

// Loop
// Monitors the Job queue and executes the function if a job is found.
func Loop(scheduler *Scheduler, f func(job *Job) error) (err error) {

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

func (scheduler *Scheduler) GetBy(filter *Filter) []Job {

	jobs := make([]Job, 0)
	if filter == nil {
		return jobs
	}

	scheduler.mutex.RLock()
	defer scheduler.mutex.RUnlock()

	useId := filter.UseIdentifier()
	useModule := filter.UseModule()
	useCluster := filter.UseCluster()
	useInterval := filter.UseInterval()

	for _, job := range scheduler.jobs {

		moduleMatch := job.Module == filter.Module
		clusterMatch := job.Cluster == filter.Cluster
		intervalMatch := job.Interval.Equals(&filter.Interval)

		if useId && (job.Identifier == filter.Identifier) {
			jobs = append(jobs, job)
			break
		} else if (useModule && moduleMatch) ||
			(useCluster && moduleMatch && clusterMatch) ||
			(useInterval && moduleMatch && clusterMatch && intervalMatch) {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func (scheduler *Scheduler) Create(job *Job) error {

	// only create a read lock for the duration we are validating
	// no other equivalent job exists within the scheduler as to
	// no interrupt parallel read tasks
	scheduler.mutex.RLock()

	found := false

	for _, jobInstance := range scheduler.jobs {
		if jobInstance.Equals(job) {
			found = true
			break
		}
	}

	if found {
		scheduler.mutex.RUnlock()
		return errors.New("identical job already exists")
	}

	scheduler.mutex.RUnlock()

	// we need to modify the jobs list, so NOW risk interrupting
	// other threads attempting to use the job list
	scheduler.mutex.Lock()
	defer scheduler.mutex.Unlock()

	scheduler.jobs = append(scheduler.jobs, *job) // create an owning copy
	return nil
}

func (scheduler *Scheduler) Delete(filter *Filter) error {

	if filter == nil {
		return errors.New("received nil filter pointer")
	}

	scheduler.mutex.Lock()
	defer scheduler.mutex.Unlock()

	useId := filter.UseIdentifier()
	useModule := filter.UseModule()
	useCluster := filter.UseCluster()
	useInterval := filter.UseInterval()

	for idx, jobInstance := range scheduler.jobs {

		moduleSame := jobInstance.Module == filter.Module
		clusterSame := jobInstance.Cluster == filter.Cluster
		intervalSame := jobInstance.Interval.Equals(&filter.Interval)

		if useId && (jobInstance.Identifier == filter.Identifier) {
			scheduler.jobs = append(scheduler.jobs[:idx], scheduler.jobs[idx+1:]...)
			break
		} else if (useModule && moduleSame) || (useCluster && moduleSame && clusterSame) || (useInterval && moduleSame && clusterSame && intervalSame) {
			scheduler.jobs = append(scheduler.jobs[:idx], scheduler.jobs[idx+1:]...)
		}
	}

	return nil
}

func (scheduler *Scheduler) GetQueue() []Job {

	scheduler.mutex.RLock()
	defer scheduler.mutex.RUnlock()

	jobs := make([]Job, len(scheduler.queue))
	for idx, job := range scheduler.queue {
		jobs[idx] = *job // make a copy
	}

	return jobs
}

func (scheduler *Scheduler) Print() {

	for _, job := range scheduler.jobs {
		fmt.Printf("├─ %s\n", job.ToString())
	}
}
