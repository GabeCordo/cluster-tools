package scheduler

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
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

		for _, job := range scheduler.jobs {

			if IsTimeToRun(&job) {
				scheduler.queue = append(scheduler.queue, &job)
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
		if len(scheduler.queue) >= 1 {
			// pop the first element of the queue (FIFO) and remove the
			// first element by slicing out the first element
			popped := scheduler.queue[0]
			scheduler.queue = scheduler.queue[1:]
			// to abide by the pattern, if the called function returns an
			// error stop the scheduler loop
			if err = f(popped); err != nil {
				break
			}
		}
		// the time till the next queue check is defined in the Scheduler config
		time.Sleep(time.Duration(scheduler.config.RefreshInterval) * time.Millisecond)
	}

	return err
}

func (scheduler Scheduler) Print() {

	for _, job := range scheduler.jobs {
		fmt.Printf("├─ %s\n", job.ToString())
	}
}
