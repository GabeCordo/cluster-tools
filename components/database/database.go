package database

import (
	"github.com/GabeCordo/etl/components/cluster"
	"time"
)

func NewDatabase() *Database {
	db := new(Database)
	db.Records = make(map[string]*Record)
	db.Configs = make(map[string]cluster.Config)
	return db
}

type RecordFactory func(identifier string) *Record

func (db *Database) CreateUsageRecord(identifier string) *Record {
	record := NewRecord()
	db.Records[identifier] = record

	return record
}

func (db *Database) GetUsageRecord(identifier string) (*Record, bool) {
	if record, found := db.Records[identifier]; found {
		return record, true
	} else {
		return nil, false
	}
}

func (db *Database) StoreUsageRecord(identifier string, data *cluster.Statistics, elpased time.Duration) bool {
	if data == nil {
		return false
	}

	db.mutex.Lock()

	record, ok := db.GetUsageRecord(identifier)
	if !ok {
		record = db.CreateUsageRecord(identifier)
	}

	db.mutex.Unlock()

	record.mutex.Lock()

	record.Head++
	entry := Entry{time.Now(), elpased, *data} //	make a copy of the stats
	record.Entries[record.Head] = entry

	record.mutex.Unlock()

	return true
}

func (db *Database) StoreClusterConfig(config cluster.Config) bool {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.Configs[config.Identifier]; found {
		return false
	}

	db.Configs[config.Identifier] = config

	return true
}

func (db *Database) ReplaceClusterConfig(config cluster.Config) bool {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, found := db.Configs[config.Identifier]; !found {
		return false
	}

	db.Configs[config.Identifier] = config

	return true
}

func (db *Database) GetClusterConfig(cluster string) (config *cluster.Config, found bool) {

	if config, found := db.Configs[cluster]; !found {
		return nil, false
	} else {
		return &config, true
	}
}
