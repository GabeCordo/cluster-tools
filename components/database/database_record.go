package database

func NewRecord() *Record {
	record := new(Record)

	record.Entries = [MaxClusterRecordSize]Entry{}
	record.Head = -1

	return record
}

func (r *Record) Empty() bool {
	return r.Head == -1
}
