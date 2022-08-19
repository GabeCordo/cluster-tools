package database

import "ETLFramework/cluster"

func NewStatisticsRecord() *StatisticRecord {
	record := new(StatisticRecord)
	record.Entries = [MaxClusterRecordSize]Entry{}
	record.Head = Empty

	return record
}

func (r StatisticRecord) Empty() bool {
	return r.Head == -1
}

func NewMonitorRecord() *MonitorRecord {
	record := new(MonitorRecord)
	record.Entries = [cluster.MaxConcurrentMonitors]Monitor{}
	record.Head = Empty

	return record
}

func (r MonitorRecord) Empty() bool {
	return r.Head == -1
}
