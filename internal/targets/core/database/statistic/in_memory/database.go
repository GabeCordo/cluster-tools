package in_memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/ScalingFunctions"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic"
)

type LocalDatabase struct {
	records map[string]map[string][]*statistic.Statistic
	mutex   sync.RWMutex
}

func NewLocalDatabase() *LocalDatabase {

	db := new(LocalDatabase)
	db.records = make(map[string]map[string][]*statistic.Statistic)

	return db
}

// Save moves *Statistic records from the statistic.LocalDatabase to the disk.
func (localDatabase *LocalDatabase) Save(path string) error {

	path = filepath.Clean(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	fileName := fmt.Sprintf("DistributedFunctions_stats_%s.json", time.Now().Format(time.RFC3339))
	outputFilePath := filepath.Join(path, fileName)

	if _, err := os.Stat(outputFilePath); os.IsExist(err) {
		return err
	}

	f, err := os.Create(outputFilePath) // #nosec G304 -- 'path' var in 'outputFilePath' is already cleaned
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(f)

	statisticBytes, _ := json.Marshal(localDatabase.records)
	_, err = f.Write(statisticBytes)
	return err
}

// Load is not implemented for the statistic.LocalDatabase.
func (localDatabase *LocalDatabase) Load(path string) (err error) {

	err = database.NotImplemented
	return err
}

// Get returns a list of *ScalingFunctions.Statistic records in the statistic.LocalDatabase.
func (localDatabase *LocalDatabase) Get(filter database.Filter) (results []*ScalingFunctions.Statistics) {

	localDatabase.mutex.RLock()
	defer localDatabase.mutex.RUnlock()

	if filter.Namespace == "" {
		return results
	}

	module, found := localDatabase.records[filter.Namespace]

	if !found {
		return results
	}

	if filter.Pipeline == "" {
		return results
	}

	records, found := module[filter.Pipeline]
	if !found {
		return results
	}

	for _, record := range records {
		results = append(results, record.Data)
	}

	return results
}

// Create adds a new *ScalingFunctions.Statistic record to the statistic.LocalDatabase.
func (localDatabase *LocalDatabase) Create(filter database.Filter, record *ScalingFunctions.Statistics) (*ScalingFunctions.Statistics, error) {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	if _, found := localDatabase.records[filter.Namespace]; !found {
		localDatabase.records[filter.Namespace] = make(map[string][]*statistic.Statistic)
	}

	s := &statistic.Statistic{
		Namespace: filter.Namespace,
		Timestamp: time.Now(),
		Pipeline:  filter.Pipeline,
		Data:      record,
	}

	if _, found := localDatabase.records[filter.Namespace][filter.Pipeline]; !found {
		statistics := make([]*statistic.Statistic, 1)
		statistics[0] = s
		localDatabase.records[filter.Namespace][filter.Pipeline] = statistics
	} else {
		localDatabase.records[filter.Namespace][filter.Pipeline] = append(localDatabase.records[filter.Namespace][filter.Pipeline], s)
	}

	return record, nil
}

// Delete removes a *ScalingFunctions.Statistic record from the statistic.LocalDatabase.
func (localDatabase *LocalDatabase) Delete(filter database.Filter) error {

	localDatabase.mutex.Lock()
	defer localDatabase.mutex.Unlock()

	if _, found := localDatabase.records[filter.Namespace]; !found {
		return errors.New("module does not exist")
	}

	delete(localDatabase.records, filter.Namespace)
	return nil
}

// Replace is not implemented for the statistic.LocalDatabase.
func (localDatabase *LocalDatabase) Replace(filter database.Filter, record *ScalingFunctions.Statistics) (err error) {

	err = database.NotImplemented
	return err
}

// Distinct is not implemented for the statistic.LocalDatabase
func (localDatabase *LocalDatabase) Distinct(filter database.Filter) (results []any, err error) {

	err = database.NotImplemented
	return results, err
}

// Print outputs the *ScalingFunctions.Statistic records inside the statistic.LocalDatabase to the console.
func (localDatabase *LocalDatabase) Print() {

	for moduleName, module := range localDatabase.records {

		fmt.Printf("├─ %s\n", moduleName)

		for supervisorName, statistics := range module {

			fmt.Printf("|   ├─ %s (num of records: %d) \n", supervisorName, len(statistics))
		}
	}
}
