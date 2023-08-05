package database

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"time"
)

func NewDatabase() *Database {
	db := new(Database)
	db.Records = make(map[string]map[string]*Record)
	db.Configs = make(map[string]map[string]cluster.Config)
	return db
}

type RecordFactory func(identifier string) *Record

func (db *Database) CreateUsageRecord(module, identifier string) *Record {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.Records[module]; !found {
		db.Records[module] = make(map[string]*Record)
	}

	record := NewRecord()
	db.Records[module][identifier] = record

	return record
}

func (db *Database) GetUsageRecord(module, identifier string) (*Record, bool) {

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	moduleMapping, moduleFound := db.Records[module]
	if !moduleFound {
		return nil, false
	}

	recordMapping, recordFound := moduleMapping[identifier]
	if !recordFound {
		return nil, false

	}

	return recordMapping, true
}

func (db *Database) StoreUsageRecord(module, identifier string, data *cluster.Statistics, elapsed time.Duration) bool {
	if data == nil {
		return false
	}

	record, ok := db.GetUsageRecord(module, identifier)
	if !ok {
		record = db.CreateUsageRecord(module, identifier)
	}

	record.mutex.Lock()

	record.Head++
	entry := Entry{time.Now(), elapsed, *data} //	make a copy of the stats
	record.Entries[record.Head] = entry

	record.mutex.Unlock()

	return true
}

func (db *Database) StoreClusterConfig(moduleName string, config cluster.Config) bool {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	/* if the module doesn't exist, since this is a create function, we can initialize a module */
	if _, found := db.Configs[moduleName]; !found {
		db.Configs[moduleName] = make(map[string]cluster.Config)
	}

	/* if the common already exists, we can't replace it, use the replace common function */
	if _, found := db.Configs[moduleName][config.Identifier]; found {
		return false
	}

	db.Configs[moduleName][config.Identifier] = config

	return true
}

func (db *Database) ReplaceClusterConfig(moduleName string, config cluster.Config) (success bool) {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	/* we're not creating a module using the replace function */
	if _, found := db.Configs[moduleName]; !found {
		return false
	}

	/* we're not creating a new common using the replace function */
	if _, found := db.Configs[moduleName][config.Identifier]; !found {
		return false
	}

	db.Configs[moduleName][config.Identifier] = config

	return true
}

func (db *Database) GetClusterConfig(moduleName, configName string) (config cluster.Config, found bool) {

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if _, found = db.Configs[moduleName]; !found {
		return cluster.Config{}, found
	}

	if _, found = db.Configs[moduleName][configName]; !found {
		return cluster.Config{}, found
	}

	config = db.Configs[moduleName][configName]
	return config, found
}

func (db *Database) DeleteModuleRecords(moduleName string) (success bool) {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.Records[moduleName]; found {
		delete(db.Records, moduleName)
	}

	if _, found := db.Configs[moduleName]; found {
		delete(db.Configs, moduleName)
	}

	return true
}
