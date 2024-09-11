package database

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type LocalJobDatabase struct {
	jobs  []interfaces.Job
	mutex sync.RWMutex
}

func NewLocalJobDatabase() *LocalJobDatabase {
	j := new(LocalJobDatabase)
	j.jobs = make([]interfaces.Job, 0)
	return j
}

// Initialize
// Copy the contents of a scheduler dump into the scheduler memory.
func (database *LocalJobDatabase) Initialize(dump *interfaces.Dump) error {

	for _, job := range dump.Jobs {
		database.jobs = append(database.jobs, job)
	}

	return nil
}

// Load
// Loads a set of static jobs defined in a yaml file into runtime.
func (database *LocalJobDatabase) Load(path string) error {

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

		dump := &interfaces.Dump{}
		if err = yaml.Unmarshal(b, dump); err != nil {
			return err
		}

		if err = database.Initialize(dump); err != nil {
			return err
		}

		return nil
	})

	// if filepath.walkdir returns an error, it will be passed here
	return err
}

// Save
// Move all jobs in the scheduler into files seperated by their modules.
func (database *LocalJobDatabase) Save(path string) error {

	if fInfo, err := os.Stat(path); os.IsNotExist(err) || (os.IsExist(err) && !fInfo.IsDir()) {
		output := fmt.Sprintf("%s is not a valid path to a directory", path)
		return errors.New(output)
	}

	// clear all the files that already existed in the folder, they should have been loaded
	// into the scheduler if the Load function was called correctly
	os.RemoveAll(path)
	os.MkdirAll(path, 0750)

	moduleSeperatedJobs := make(map[string][]interfaces.Job)

	// order all the jobs by the module they belong to
	for _, job := range database.jobs {
		if _, found := moduleSeperatedJobs[job.Module]; !found {
			moduleSeperatedJobs[job.Module] = make([]interfaces.Job, 0)
		}
		moduleSeperatedJobs[job.Module] = append(moduleSeperatedJobs[job.Module], job)
	}

	// save each module's job into its own file
	for module, jobs := range moduleSeperatedJobs {

		filePath := fmt.Sprintf("%s/schedule_%s.yml", path, module)
		dump := &interfaces.Dump{Jobs: jobs}
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

func (database *LocalJobDatabase) GetAll() ([]interfaces.Job, error) {
	return database.jobs, nil
}

func (database *LocalJobDatabase) GetBy(filter *interfaces.Filter) ([]interfaces.Job, error) {

	jobs := make([]interfaces.Job, 0)
	if filter == nil {
		return jobs, nil
	}

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	useId := filter.UseIdentifier()
	useModule := filter.UseModule()
	useCluster := filter.UseCluster()
	useInterval := filter.UseInterval()

	for _, job := range database.jobs {

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

	return jobs, nil
}

func (database *LocalJobDatabase) Create(job *interfaces.Job) error {

	// only create a read lock for the duration we are validating
	// no other equivalent job exists within the scheduler as to
	// no interrupt parallel read tasks
	database.mutex.RLock()

	found := false

	for _, jobInstance := range database.jobs {
		if jobInstance.Equals(job) {
			found = true
			break
		}
	}

	if found {
		database.mutex.RUnlock()
		return errors.New("identical job already exists")
	}

	database.mutex.RUnlock()

	// we need to modify the jobs list, so NOW risk interrupting
	// other threads attempting to use the job list
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.jobs = append(database.jobs, *job) // create an owning copy
	return nil
}

func (database *LocalJobDatabase) Delete(filter *interfaces.Filter) error {

	if filter == nil {
		return errors.New("received nil filter pointer")
	}

	database.mutex.Lock()
	defer database.mutex.Unlock()

	useId := filter.UseIdentifier()
	useModule := filter.UseModule()
	useCluster := filter.UseCluster()
	useInterval := filter.UseInterval()

	for idx, jobInstance := range database.jobs {

		moduleSame := jobInstance.Module == filter.Module
		clusterSame := jobInstance.Cluster == filter.Cluster
		intervalSame := jobInstance.Interval.Equals(&filter.Interval)

		if useId && (jobInstance.Identifier == filter.Identifier) {
			database.jobs = append(database.jobs[:idx], database.jobs[idx+1:]...)
			break
		} else if (useModule && moduleSame) || (useCluster && moduleSame && clusterSame) || (useInterval && moduleSame && clusterSame && intervalSame) {
			database.jobs = append(database.jobs[:idx], database.jobs[idx+1:]...)
		}
	}

	return nil
}

func (database *LocalJobDatabase) Print() {

	for _, job := range database.jobs {
		fmt.Printf("├─ %s\n", job.ToString())
	}
}
