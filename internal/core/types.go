package core

type ThreadType uint8

const (
	HttpClient ThreadType = iota
	HttpProcessor
	Processor
	Supervisor
	Database
	Messenger
	Cache
	Scheduler
	Undefined
)

func (threadType ThreadType) ToString() string {
	switch threadType {
	case HttpClient:
		return "HTTP-CLIENT"
	case HttpProcessor:
		return "HTTP-PROCESSOR"
	case Processor:
		return "PROCESSOR"
	case Supervisor:
		return "SUPERVISOR"
	case Messenger:
		return "MESSENGER"
	case Database:
		return "DATABASE"
	case Cache:
		return "CACHE"
	case Scheduler:
		return "SCHEDULER"
	default:
		return "-"
	}
}
