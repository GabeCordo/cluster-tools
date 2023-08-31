package messenger

import "sync"

type MessagePriority uint

const (
	Log MessagePriority = iota
	Warning
	Fatal
)

func (priority MessagePriority) ToString() string {
	if priority == Log {
		return "-"
	} else if priority == Warning {
		return "?"
	} else {
		return "!"
	}
}

type Cluster struct {
	supervisors map[uint64][]string
	mutex       sync.RWMutex
}

func NewCluster() *Cluster {
	instance := new(Cluster)
	instance.supervisors = make(map[uint64][]string)
	return instance
}

type Module struct {
	clusters map[string]*Cluster
	mutex    sync.RWMutex
}

func NewModule() *Module {
	instance := new(Module)
	instance.clusters = make(map[string]*Cluster)
	return instance
}

type Messenger struct {
	enabled struct {
		logging bool
		smtp    bool
	}
	logging struct {
		directory string
	}
	smtp struct {
		endpoint    Endpoint
		credentials Credentials
		receivers   map[string][]string
	}

	modules map[string]*Module
	mutex   sync.RWMutex
}

func NewMessenger(enableLogging, enableSmtp bool) *Messenger {
	messenger := new(Messenger)
	messenger.modules = make(map[string]*Module)

	messenger.enabled.logging = enableLogging
	messenger.enabled.smtp = enableSmtp

	return messenger
}
