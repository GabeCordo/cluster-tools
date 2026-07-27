package in_memory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/job"

	"gopkg.in/yaml.v3"
)

const defaultFilePerm = 0600

type LocalDatabase struct {
	jobs  []*job.Job
	mutex sync.RWMutex
}

func NewLocalJobDatabase() *LocalDatabase {
	j := new(LocalDatabase)
	j.jobs = make([]*job.Job, 0)
	return j
}

// Initialize copies the contents of a scheduler dump into the scheduler memory.
func (localDatabase *LocalDatabase) Initialize(dump job.Dump) error {

	for _, job := range dump.Jobs {
		localDatabase.jobs = append(localDatabase.jobs, job)
	}

	return nil
}

// Load moves Job records from disk to the job.LocalDatabase.
func (localDatabase *LocalDatabase) Load(path string) error {

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

		path = filepath.Clean(path)
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		dump := job.Dump{}
		if err = yaml.Unmarshal(b, &dump); err != nil {
			return err
		}

		if err = localDatabase.Initialize(dump); err != nil {
			return err
		}

		return nil
	})

	// if filepath.walkdir returns an error, it will be passed here
	return err
}

// Save moves Job records from job.LocalDatabase to the disk.
func (localDatabase *LocalDatabase) Save(path string) error {

	if fInfo, err := os.Stat(path); os.IsNotExist(err) || (os.IsExist(err) && !fInfo.IsDir()) {
		output := fmt.Sprintf("%s is not a valid path to a directory", path)
		return errors.New(output)
	}

	// clear all the files that already existed in the folder, they should have been loaded
	// into the scheduler if the Load function was called correctly
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path, defaultFilePerm)
	if err != nil {
		return err
	}

	moduleSeperatedJobs := make(map[string][]*job.Job)

	// order all the jobs by the module they belong to
	for _, j := range localDatabase.jobs {
		if _, found := moduleSeperatedJobs[j.Namespace]; !found {
			moduleSeperatedJobs[j.Namespace] = make([]*job.Job, 0)
		}
		moduleSeperatedJobs[j.Namespace] = append(moduleSeperatedJobs[j.Namespace], j)
	}

	// save each module's job into its own file
	for module, jobs := range moduleSeperatedJobs {

		fileName := fmt.Sprintf("schedule_%s.yml", module)
		filePath := filepath.Join(path, fileName)
		dump := &job.Dump{Jobs: jobs}
		b, err := yaml.Marshal(dump)
		if err != nil {
			output := fmt.Sprintf("failed to turn jobs into dump file %s", err.Error())
			return errors.New(output)
		}
		if err = os.WriteFile(filePath, b, defaultFilePerm); err != nil {
			output := fmt.Sprintf("failed to write dump to file %s", err.Error())
			return errors.New(output)
		}
	}

	return nil
}

// Get retrieves a Job record from the job.LocalDatabase.
func (localDatabase *LocalDatabase) Get(filter database.Filter) (jobs []*job.Job) {

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	if filter.IsEmpty() {
		for _, j := range localDatabase.jobs {
			jobs = append(jobs, j)
		}
		return jobs
	}

	useId := filter.UseIdentifier()
	useModule := filter.UseNamespace()
	useCluster := filter.UsePipeline()
	useInterval := filter.UseInterval()

	for _, j := range localDatabase.jobs {

		moduleMatch := j.Namespace == filter.Namespace
		clusterMatch := j.Pipeline == filter.Pipeline
		intervalMatch := j.Interval.Equals(&filter.Interval)

		if useId && (j.Identifier == filter.Identifier) {
			jobs = append(jobs, j)
			break
		} else if (useModule && moduleMatch) ||
			(useCluster && moduleMatch && clusterMatch) ||
			(useInterval && moduleMatch && clusterMatch && intervalMatch) {
			jobs = append(jobs, j)
		}
	}

	return jobs
}

// Create adds a *Job record to the job.LocalDatabase.
func (localDatabase *LocalDatabase) Create(filter database.Filter, record *job.Job) (string, error) {

	// only create a read lock for the duration we are validating
	// no other equivalent job exists within the scheduler as to
	// no interrupt parallel read tasks
	localDatabase.mutex.RLock()

	found := false

	for _, jobInstance := range localDatabase.jobs {
		if jobInstance.Equals(record) {
			found = true
			break
		}
	}

	if found {
		localDatabase.mutex.RUnlock()
		return "", errors.New("identical job already exists")
	}

	localDatabase.mutex.RUnlock()

	// we need to modify the jobs list, so NOW risk interrupting
	// other thread attempting to use the job list
	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	localDatabase.jobs = append(localDatabase.jobs, record)
	return record.Identifier, nil
}

// Delete removes a *Job from the job.LocalDatabase.
func (localDatabase *LocalDatabase) Delete(filter database.Filter) error {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	useId := filter.UseIdentifier()
	useModule := filter.UseNamespace()
	useCluster := filter.UsePipeline()
	useInterval := filter.UseInterval()

	for idx, jobInstance := range localDatabase.jobs {

		moduleSame := jobInstance.Namespace == filter.Namespace
		clusterSame := jobInstance.Pipeline == filter.Pipeline
		intervalSame := jobInstance.Interval.Equals(&filter.Interval)

		if useId && (jobInstance.Identifier == filter.Identifier) {
			localDatabase.jobs = append(localDatabase.jobs[:idx], localDatabase.jobs[idx+1:]...)
			break
		} else if (useModule && moduleSame) || (useCluster && moduleSame && clusterSame) || (useInterval && moduleSame && clusterSame && intervalSame) {
			localDatabase.jobs = append(localDatabase.jobs[:idx], localDatabase.jobs[idx+1:]...)
		}
	}

	return nil
}

// Replace is not implemented for the job.LocalDatabase.
func (localDatabase *LocalDatabase) Replace(filter database.Filter, record *job.Job) (err error) {

	err = database.NotImplemented
	return err
}

// Print outputs the *Job records in the job.LocalDatabase to the console.
func (localDatabase *LocalDatabase) Print() {

	for _, job := range localDatabase.jobs {
		fmt.Printf("├─ %s\n", job.ToString())
	}
}
