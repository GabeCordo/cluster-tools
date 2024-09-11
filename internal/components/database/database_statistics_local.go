package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type LocalStatisticDatabase struct {
	records map[string]map[string][]Statistic
	mutex   sync.RWMutex
}

func NewLocalStatisticDatabase() *LocalStatisticDatabase {

	db := new(LocalStatisticDatabase)
	db.records = make(map[string]map[string][]Statistic)

	return db
}

func (db *LocalStatisticDatabase) Save(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	outputFilePath := fmt.Sprintf("%s/etl_stats_%s.json", path, time.Now().Format(time.RFC3339))

	if _, err := os.Stat(outputFilePath); os.IsExist(err) {
		return err
	}

	f, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	statisticBytes, _ := json.Marshal(db.records)
	f.Write(statisticBytes)

	return nil
}

type StatisticFilter struct {
	Module  string
	Cluster string
	Verbose bool
}

func (db *LocalStatisticDatabase) Get(filter StatisticFilter) (records []Statistic, err error) {

	records = nil
	err = nil

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if filter.Module == "" {
		err = errors.New("module cannot be empty")
		return records, err
	}

	module, found := db.records[filter.Module]

	if !found {
		err = errors.New("module does not exist")
		return records, err
	}

	if filter.Cluster == "" {
		err = errors.New("config cannot be empty")
		return records, err
	}

	records, found = module[filter.Cluster]
	if !found {
		err = errors.New("cluster does not exist")
		return records, err
	}

	return records, err
}

func (db *LocalStatisticDatabase) Create(moduleId, clusterId string, statistic Statistic) (err error) {

	err = nil

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.records[moduleId]; !found {
		db.records[moduleId] = make(map[string][]Statistic)
	}

	if _, found := db.records[moduleId][clusterId]; !found {
		statistics := make([]Statistic, 1)
		statistics[0] = statistic
		db.records[moduleId][clusterId] = statistics
	} else {
		db.records[moduleId][clusterId] = append(db.records[moduleId][clusterId], statistic)
	}

	return err
}

func (db *LocalStatisticDatabase) Delete(moduleId string) (err error) {

	err = nil

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.records[moduleId]; !found {
		err = errors.New("module does not exist")
		return err
	}

	delete(db.records, moduleId)
	return err
}

func (db *LocalStatisticDatabase) Print() {

	for moduleName, module := range db.records {

		fmt.Printf("├─ %s\n", moduleName)

		for supervisorName, statistics := range module {

			fmt.Printf("|   ├─ %s (num of records: %d) \n", supervisorName, len(statistics))
		}
	}
}
