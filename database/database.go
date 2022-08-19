package database

import (
	"ETLFramework/cluster"
	"time"
)

func NewDatabase() *Database {
	db := new(Database)

	db.statistics = make(map[string]*StatisticRecord)
	db.monitors = make(map[string]*MonitorRecord)

	return db
}

type RecordFactory func(identifier string) *Record

func (db *Database) CreateStatisticRecord(identifier string) *StatisticRecord {
	record := NewStatisticsRecord()
	db.statistics[identifier] = record

	return record
}

func (db *Database) CreateMonitorRecord(identifier string) *MonitorRecord {
	record := NewMonitorRecord()
	db.monitors[identifier] = record

	return record
}

func (db *Database) GetStatisticRecord(identifier string) (*StatisticRecord, bool) {
	if record, found := db.statistics[identifier]; found {
		return record, true
	} else {
		return nil, false
	}
}

func (db *Database) GetMonitorRecord(identifier string) (*MonitorRecord, bool) {
	if record, found := db.monitors[identifier]; found {
		return record, true
	} else {
		return nil, false
	}
}

func (db *Database) StoreStatistic(identifier string, data *cluster.Statistics, elpased time.Duration) bool {
	if data == nil {
		return false
	}

	db.mutex.Lock()

	record, ok := db.GetStatisticRecord(identifier)
	if !ok {
		record = db.CreateStatisticRecord(identifier)
	}

	db.mutex.Unlock()

	record.mutex.Lock()

	record.Head++
	entry := Entry{time.Now(), elpased, *data} //	make a copy of the stats
	record.Entries[record.Head] = entry

	record.mutex.Unlock()

	return true
}

func (db *Database) StoreMonitor(identifier string, data *Monitor) int32 {
	if data == nil {
		return -1
	}

	db.mutex.Lock()

	record, ok := db.GetMonitorRecord(identifier)
	if !ok {
		record = db.CreateMonitorRecord(identifier)
	}

	db.mutex.Unlock()

	record.mutex.Lock()

	record.Head++
	entry := *data // make a copy of the monitor
	record.Entries[record.Head] = entry

	record.mutex.Unlock()

	return interface{}(record.Head).(int32)
}

func (db *Database) ModifyMonitor(id MonitorId, data *Monitor) bool {
	if data == nil {
		return false
	}

	if record, found := db.monitors[id.Identifier]; found {
		if record.Head < id.Id {
			return false
		}

		record.mutex.Lock()

		monitor := &record.Entries[id.Id] // create a pointer to the current Monitor record
		monitor.NumTransformRoutines = data.NumTransformRoutines
		monitor.NumExtractRoutines = data.NumExtractRoutines
		monitor.NumLoadRoutines = data.NumLoadRoutines

		record.mutex.Unlock()
	}

	return true
}
