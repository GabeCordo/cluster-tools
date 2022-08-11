package database

func (r Record) Empty() bool {
	return r.Head == -1
}
