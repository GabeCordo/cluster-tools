package database

import (
	"ETLFramework/cluster"
	"time"
)

func NewDatabase() *Database {
	db := new(Database)
	db.Records = make(map[string]*Record)

	return db
}

func (db *Database) CreateRecord(identifier string) *Record {
	// the record does not exist and should be initialized / created
	record := new(Record)
	record.Entries = [MaxClusterRecordSize]Entry{}
	record.Head = -1

	db.Records[identifier] = record // pass by copy of pointer

	return record
}

func (db *Database) GetRecord(identifier string) (*Record, bool) {
	if record, found := db.Records[identifier]; found {
		return record, true
	} else {
		return nil, false
	}
}

func (db *Database) Store(identifier string, data cluster.Response) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	record, ok := db.GetRecord(identifier)
	if !ok {
		record = db.CreateRecord(identifier)
	}

	record.Head++
	entry := Entry{time.Now(), data.LapsedTime, data.Stats.NumProvisionedTransformRoutes, data.Stats.NumProvisionedLoadRoutines}
	record.Entries[record.Head] = entry
}
