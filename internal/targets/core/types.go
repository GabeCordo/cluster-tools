package core

type ThreadType uint8

const (
	RestAPI ThreadType = iota
	Socket
	Processor
	Runner
	Database
	Messenger
	Cache
	Scheduler
	Undefined
)

func (threadType ThreadType) ToString() string {
	switch threadType {
	case RestAPI:
		return "HTTP-CLIENT"
	case Socket:
		return "TLS-SOCKET"
	case Processor:
		return "PROCESSOR"
	case Runner:
		return "RUNNER"
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
