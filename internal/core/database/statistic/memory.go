package statistic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/FortifiedCode/flock/internal/core/database"
)

type LocalStatisticDatabase struct {
	records map[string]map[string][]Wrapper
	mutex   sync.RWMutex
}

func NewLocalStatisticDatabase() *LocalStatisticDatabase {

	db := new(LocalStatisticDatabase)
	db.records = make(map[string]map[string][]Wrapper)

	return db
}

func (db *LocalStatisticDatabase) Save(path string) error {

	path = filepath.Clean(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	fileName := fmt.Sprintf("flock_stats_%s.json", time.Now().Format(time.RFC3339))
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

	statisticBytes, _ := json.Marshal(db.records)
	_, err = f.Write(statisticBytes)
	return err
}

func (db *LocalStatisticDatabase) Load(path string) error {
	panic("implement me")
}

type Filter struct {
	Module  string
	Cluster string
	Verbose bool
}

func (db *LocalStatisticDatabase) Get(filter database.Filter) []any {

	// (records []Wrapper, err error)
	results := make([]any, 0)

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if filter.Namespace == "" {
		return results
	}

	module, found := db.records[filter.Namespace]

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
		results = append(results, record.Stats)
	}

	return results
}

func (db *LocalStatisticDatabase) Create(filter database.Filter, record any) (any, error) {

	// old: moduleId, clusterId string, statistic Wrapper

	statistic, ok := record.(Wrapper)
	if !ok {
		return nil, errors.New("invalid record type")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.records[filter.Namespace]; !found {
		db.records[filter.Namespace] = make(map[string][]Wrapper)
	}

	if _, found := db.records[filter.Namespace][filter.Pipeline]; !found {
		statistics := make([]Wrapper, 1)
		statistics[0] = statistic
		db.records[filter.Namespace][filter.Pipeline] = statistics
	} else {
		db.records[filter.Namespace][filter.Pipeline] = append(db.records[filter.Namespace][filter.Pipeline], statistic)
	}

	return filter.Pipeline, nil
}

func (db *LocalStatisticDatabase) Delete(filter database.Filter) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.records[filter.Namespace]; !found {
		return errors.New("module does not exist")
	}

	delete(db.records, filter.Namespace)
	return nil
}

func (db *LocalStatisticDatabase) Replace(filter database.Filter, record any) error {
	panic("implement me")
}

func (db *LocalStatisticDatabase) Print() {

	for moduleName, module := range db.records {

		fmt.Printf("├─ %s\n", moduleName)

		for supervisorName, statistics := range module {

			fmt.Printf("|   ├─ %s (num of records: %d) \n", supervisorName, len(statistics))
		}
	}
}
