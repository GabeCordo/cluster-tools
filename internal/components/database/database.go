package database

type Database interface {
	Save(path string) error
	Load(path string) error
	Print()
}
