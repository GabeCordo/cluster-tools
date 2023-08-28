package common

type InterruptEvent uint8

const (
	Shutdown InterruptEvent = 0
	Panic                   = 1
)

type Module uint8

const (
	HttpClient Module = iota
	HttpProcessor
	Database
	Processor
	Supervisor
	Messenger
	Cache
)

type Thread interface {
	Setup()
	Start()
	Teardown()
}
