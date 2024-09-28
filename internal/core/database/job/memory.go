package job

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type LocalJobDatabase struct {
	jobs  []Job
	mutex sync.RWMutex
}

func NewLocalJobDatabase() *LocalJobDatabase {
	j := new(LocalJobDatabase)
	j.jobs = make([]Job, 0)
	return j
}

// Initialize
// Copy the contents of a scheduler dump into the scheduler memory.
func (database *LocalJobDatabase) Initialize(dump *Dump) error {

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

		dump := &Dump{}
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

	moduleSeperatedJobs := make(map[string][]Job)

	// order all the jobs by the module they belong to
	for _, job := range database.jobs {
		if _, found := moduleSeperatedJobs[job.Namespace]; !found {
			moduleSeperatedJobs[job.Namespace] = make([]Job, 0)
		}
		moduleSeperatedJobs[job.Namespace] = append(moduleSeperatedJobs[job.Namespace], job)
	}

	// save each module's job into its own file
	for module, jobs := range moduleSeperatedJobs {

		filePath := fmt.Sprintf("%s/schedule_%s.yml", path, module)
		dump := &Dump{Jobs: jobs}
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

func (database *LocalJobDatabase) Get(filter database.Filter) []any {

	jobs := make([]any, 0)

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	if filter.IsEmpty() {
		for _, job := range database.jobs {
			jobs = append(jobs, job)
		}
		return jobs
	}

	useId := filter.UseIdentifier()
	useModule := filter.UseNamespace()
	useCluster := filter.UsePipeline()
	useInterval := filter.UseInterval()

	for _, job := range database.jobs {

		moduleMatch := job.Namespace == filter.Namespace
		clusterMatch := job.Pipeline == filter.Pipeline
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

func (database *LocalJobDatabase) Create(filter database.Filter, record any) (any, error) {

	job, ok := record.(*Job)
	if !ok {
		return nil, errors.New("invalid record type")
	}

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
		return nil, errors.New("identical job already exists")
	}

	database.mutex.RUnlock()

	// we need to modify the jobs list, so NOW risk interrupting
	// other thread attempting to use the job list
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.jobs = append(database.jobs, *job) // create an owning copy
	return job.Identifier, nil
}

func (database *LocalJobDatabase) Delete(filter database.Filter) error {

	database.mutex.Lock()
	defer database.mutex.Unlock()

	useId := filter.UseIdentifier()
	useModule := filter.UseNamespace()
	useCluster := filter.UsePipeline()
	useInterval := filter.UseInterval()

	for idx, jobInstance := range database.jobs {

		moduleSame := jobInstance.Namespace == filter.Namespace
		clusterSame := jobInstance.Pipeline == filter.Pipeline
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

func (database *LocalJobDatabase) Replace(filter database.Filter, record any) error {
	panic("not implemented")
}

func (database *LocalJobDatabase) Print() {

	for _, job := range database.jobs {
		fmt.Printf("├─ %s\n", job.ToString())
	}
}
