package messenger

import (
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"strings"
	"sync"
	"time"
)

type MessagePriority string

const (
	Normal  MessagePriority = "normal"
	Warning                 = "warning"
	Fatal                   = "fatal"
	Any                     = ""
)

func (priority MessagePriority) Shortform() string {
	if priority == Normal {
		return "-"
	} else if priority == Warning {
		return "?"
	} else if priority == Fatal {
		return "!"
	} else {
		return ""
	}
}

func PriorityFromShortform(shortform string) MessagePriority {
	if shortform == "-" {
		return Normal
	} else if shortform == "?" {
		return Warning
	} else if shortform == "!" {
		return Fatal
	} else {
		return Any
	}
}

type Log struct {
	Timestamp time.Time
	Priority  MessagePriority
	Message   string
}

func NewLog() *Log {
	log := new(Log)
	return log
}

type LogFile struct {
	Logs      []*Log
	NumOfLogs int
}

func NewLogFile(bytes []byte) *LogFile {
	instance := new(LogFile)

	logs := strings.Split(string(bytes), "\n")
	instance.NumOfLogs = len(logs) - 1 // there will always be an empty split due to the final \n
	instance.Logs = make([]*Log, instance.NumOfLogs)

	for idx, logStr := range logs {
		log := NewLog()
		if err := log.Parse(logStr); err == nil {
			instance.Logs[idx] = log
		}
	}

	return instance
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
	smtp interfaces.SmtpRecord

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
