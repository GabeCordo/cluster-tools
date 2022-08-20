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

type RecordFactory func(identifier string) *Record

func (db *Database) CreateRecord(identifier string) *Record {
	record := NewRecord()
	db.Records[identifier] = record

	return record
}

func (db *Database) GetRecord(identifier string) (*Record, bool) {
	if record, found := db.Records[identifier]; found {
		return record, true
	} else {
		return nil, false
	}
}

func (db *Database) Store(identifier string, data *cluster.Statistics, elpased time.Duration) bool {
	if data == nil {
		return false
	}

	db.mutex.Lock()

	record, ok := db.GetRecord(identifier)
	if !ok {
		record = db.CreateRecord(identifier)
	}

	db.mutex.Unlock()

	record.mutex.Lock()

	record.Head++
	entry := Entry{time.Now(), elpased, *data} //	make a copy of the stats
	record.Entries[record.Head] = entry

	record.mutex.Unlock()

	return true
}
